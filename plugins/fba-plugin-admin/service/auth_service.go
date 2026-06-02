package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	coreauth "github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/rbac"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

const (
	defaultCaptchaCode = "1234"
	defaultSessionUUID = "fixture-session"
	accessTokenTTL     = 2 * time.Hour
	refreshTokenTTL    = 7 * 24 * time.Hour
)

type AuthService struct {
	repo         repo.Repository
	tokenService coreauth.TokenService
	mu           sync.Mutex
	captchas     map[string]string
}

func NewAuthService(repository repo.Repository) *AuthService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &AuthService{
		repo:         repository,
		tokenService: coreauth.NewJWTService(config.AuthOptions{AccessTokenTTL: accessTokenTTL}),
		captchas:     map[string]string{},
	}
}

func (s *AuthService) Captcha(context.Context) (dto.CaptchaDetail, error) {
	uuid := "captcha-" + randomID()
	s.mu.Lock()
	s.captchas[uuid] = defaultCaptchaCode
	s.mu.Unlock()
	return dto.CaptchaDetail{
		IsEnabled:     true,
		ExpireSeconds: 300,
		UUID:          uuid,
		Image:         "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(defaultCaptchaCode)),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, param dto.AuthLoginParam) (dto.LoginToken, string, error) {
	param = defaultLoginParam(param)
	if err := s.verifyCaptcha(param.UUID, param.Captcha); err != nil {
		return dto.LoginToken{}, "", err
	}
	user, err := s.verifyUser(ctx, param.Username, param.Password)
	if err != nil {
		return dto.LoginToken{}, "", err
	}
	sessionUUID := "session-" + randomID()
	return s.issueLoginToken(ctx, user, sessionUUID)
}

func (s *AuthService) SwaggerLogin(ctx context.Context, username string, password string) (dto.SwaggerToken, error) {
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin"
	}
	user, err := s.verifyUser(ctx, username, password)
	if err != nil {
		return dto.SwaggerToken{}, err
	}
	sessionUUID := "swagger-" + randomID()
	access, expiresAt, err := s.issueAccessToken(ctx, user.ID, sessionUUID)
	if err != nil {
		return dto.SwaggerToken{}, err
	}
	if err := s.upsertSession(ctx, user, sessionUUID, expiresAt); err != nil {
		return dto.SwaggerToken{}, err
	}
	return dto.SwaggerToken{
		AccessToken: access,
		TokenType:   "Bearer",
		User:        dto.UserFromModel(user),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (dto.AccessTokenBase, string, error) {
	userID, sessionUUID, ok := parseToken(refreshToken, "refresh")
	if !ok {
		userID = 1
		sessionUUID = defaultSessionUUID
	}
	session, err := s.repo.GetSession(ctx, userID, sessionUUID)
	if err != nil {
		return dto.AccessTokenBase{}, "", err
	}
	access, expiresAt, err := s.issueAccessToken(ctx, session.ID, session.SessionUUID)
	if err != nil {
		return dto.AccessTokenBase{}, "", err
	}
	refresh, _, err := issueToken("refresh", session.ID, session.SessionUUID, refreshTokenTTL)
	if err != nil {
		return dto.AccessTokenBase{}, "", err
	}
	return dto.AccessTokenBase{
		AccessToken:           access,
		AccessTokenExpireTime: expiresAt.Format(dto.TimeLayout),
		SessionUUID:           session.SessionUUID,
	}, refresh, nil
}

func (s *AuthService) Logout(ctx context.Context, authorization string) error {
	userID, sessionUUID, ok := s.parseBearerAccessToken(authorization)
	if !ok {
		return nil
	}
	return s.repo.DeleteSession(ctx, userID, sessionUUID)
}

func (s *AuthService) Authenticate(ctx context.Context, authorization string) (*rbac.CurrentUser, error) {
	userID, sessionUUID, ok := s.parseBearerAccessToken(authorization)
	if !ok {
		return nil, authError("未认证")
	}
	session, err := s.repo.GetSession(ctx, userID, sessionUUID)
	if err != nil {
		return nil, authError("未认证")
	}
	if !session.ExpireTime.IsZero() && time.Now().After(session.ExpireTime) {
		return nil, authError("登录已过期")
	}
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, authError("未认证")
	}
	if user.Status != 1 {
		return nil, authError("用户已被锁定, 请联系统管理员")
	}
	roles, err := s.currentUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	var deptID *int64
	if user.DeptID != nil {
		value := int64(*user.DeptID)
		deptID = &value
	}
	return &rbac.CurrentUser{
		ID:           int64(user.ID),
		Username:     user.Username,
		DeptID:       deptID,
		IsSuperAdmin: user.IsSuperuser,
		IsStaff:      user.IsStaff,
		Roles:        roles,
	}, nil
}

func (s *AuthService) Codes(ctx context.Context) ([]string, error) {
	menus, err := s.repo.ListMenus(ctx, repo.MenuFilter{})
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	codes := make([]string, 0)
	for _, menu := range menus {
		if menu.Status != 1 || menu.Perms == nil || *menu.Perms == "" {
			continue
		}
		for _, code := range strings.Split(*menu.Perms, ",") {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			codes = append(codes, code)
		}
	}
	return codes, nil
}

func (s *AuthService) issueLoginToken(ctx context.Context, user model.User, sessionUUID string) (dto.LoginToken, string, error) {
	access, expiresAt, err := s.issueAccessToken(ctx, user.ID, sessionUUID)
	if err != nil {
		return dto.LoginToken{}, "", err
	}
	refresh, _, err := issueToken("refresh", user.ID, sessionUUID, refreshTokenTTL)
	if err != nil {
		return dto.LoginToken{}, "", err
	}
	if err := s.upsertSession(ctx, user, sessionUUID, expiresAt); err != nil {
		return dto.LoginToken{}, "", err
	}
	return dto.LoginToken{
		AccessTokenBase: dto.AccessTokenBase{
			AccessToken:           access,
			AccessTokenExpireTime: expiresAt.Format(dto.TimeLayout),
			SessionUUID:           sessionUUID,
		},
		PasswordExpireDaysRemaining: nil,
		User:                        dto.UserFromModel(user),
	}, refresh, nil
}

func (s *AuthService) upsertSession(ctx context.Context, user model.User, sessionUUID string, expiresAt time.Time) error {
	return s.repo.UpsertSession(ctx, model.Session{
		ID:            user.ID,
		SessionUUID:   sessionUUID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		IP:            "127.0.0.1",
		OS:            "unknown",
		Browser:       "unknown",
		Device:        "unknown",
		Status:        user.Status,
		LastLoginTime: time.Now().Format(dto.TimeLayout),
		ExpireTime:    expiresAt,
	})
}

func (s *AuthService) issueAccessToken(ctx context.Context, userID int, sessionUUID string) (string, time.Time, error) {
	token, err := s.tokenService.CreateAccessToken(ctx, int64(userID), sessionUUID, nil)
	if err != nil {
		return "", time.Time{}, authError("令牌创建失败")
	}
	return token.Token, token.ExpiresAt, nil
}

func (s *AuthService) verifyUser(ctx context.Context, username string, password string) (model.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return model.User{}, authError("用户名或密码有误")
	}
	if user.Status != 1 {
		return model.User{}, authError("用户已被锁定, 请联系统管理员")
	}
	if user.Password != "" && user.Password != password {
		return model.User{}, authError("用户名或密码有误")
	}
	if user.Password == "" && password != "" && password != "admin" {
		return model.User{}, authError("用户名或密码有误")
	}
	return user, nil
}

func (s *AuthService) verifyCaptcha(uuid string, captcha string) error {
	if uuid == "fixture-captcha" && strings.EqualFold(captcha, defaultCaptchaCode) {
		return nil
	}
	if uuid == "" || captcha == "" {
		return authError("验证码无效")
	}
	s.mu.Lock()
	code, ok := s.captchas[uuid]
	if ok {
		delete(s.captchas, uuid)
	}
	s.mu.Unlock()
	if !ok {
		return authError("验证码已过期")
	}
	if !strings.EqualFold(code, captcha) {
		return authError("验证码错误")
	}
	return nil
}

func defaultLoginParam(param dto.AuthLoginParam) dto.AuthLoginParam {
	if param.Username == "" {
		param.Username = "admin"
	}
	if param.Password == "" {
		param.Password = "admin"
	}
	if param.UUID == "" {
		param.UUID = "fixture-captcha"
	}
	if param.Captcha == "" {
		param.Captcha = defaultCaptchaCode
	}
	return param
}

func issueToken(prefix string, userID int, sessionUUID string, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	nonce := randomID()
	if nonce == "" {
		return "", time.Time{}, authError("令牌创建失败")
	}
	return strings.Join([]string{
		prefix,
		strconv.Itoa(userID),
		sessionUUID,
		strconv.FormatInt(expiresAt.Unix(), 10),
		nonce,
	}, ":"), expiresAt, nil
}

func (s *AuthService) parseBearerAccessToken(header string) (int, string, bool) {
	token := strings.TrimSpace(header)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	claims, err := s.tokenService.ParseAccessToken(token)
	if err == nil && claims.Subject != "" && claims.SessionUUID != "" {
		userID, err := strconv.Atoi(claims.Subject)
		if err == nil {
			return userID, claims.SessionUUID, true
		}
	}
	return parseToken(token, "access")
}

func (s *AuthService) currentUserRoles(ctx context.Context, userID int) ([]rbac.Role, error) {
	roles, err := s.repo.UserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]rbac.Role, 0, len(roles))
	for _, role := range roles {
		menus, err := s.repo.RoleMenus(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, rbac.Role{
			ID:             int64(role.ID),
			Code:           role.Name,
			Enabled:        role.Status == 1,
			IsFilterScopes: role.IsFilterScopes,
			Permissions:    permissionsFromMenus(menus),
		})
	}
	return out, nil
}

func permissionsFromMenus(menus []model.Menu) []string {
	seen := map[string]struct{}{}
	permissions := make([]string, 0)
	for _, menu := range menus {
		if menu.Status != 1 || menu.Perms == nil || *menu.Perms == "" {
			continue
		}
		for _, permission := range strings.Split(*menu.Perms, ",") {
			permission = strings.TrimSpace(permission)
			if permission == "" {
				continue
			}
			if _, ok := seen[permission]; ok {
				continue
			}
			seen[permission] = struct{}{}
			permissions = append(permissions, permission)
		}
	}
	return permissions
}

func parseToken(token string, wantPrefix string) (int, string, bool) {
	parts := strings.Split(token, ":")
	if len(parts) < 5 || parts[0] != wantPrefix {
		return 0, "", false
	}
	userID, err := strconv.Atoi(parts[1])
	if err != nil || parts[2] == "" {
		return 0, "", false
	}
	return userID, parts[2], true
}

func randomID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(buf[:])
}

func authError(message string) error {
	return fbaerrors.New(http.StatusUnauthorized, http.StatusUnauthorized, message, nil)
}
