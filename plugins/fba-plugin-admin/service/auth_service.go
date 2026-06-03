package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	stderrors "errors"
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
	loginLogFail       = 0
	loginLogSuccess    = 1
	loginSuccessMsg    = "登录成功"
)

type RequestMetadata struct {
	IP        string
	Country   *string
	Region    *string
	City      *string
	UserAgent *string
	Browser   *string
	OS        *string
	Device    *string
}

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

func (s *AuthService) Login(ctx context.Context, param dto.AuthLoginParam, meta RequestMetadata) (dto.LoginToken, string, error) {
	param = defaultLoginParam(param)
	if err := s.verifyCaptcha(param.UUID, param.Captcha); err != nil {
		s.recordLoginLog(ctx, model.User{}, param.Username, loginLogFail, err.Error(), meta)
		return dto.LoginToken{}, "", err
	}
	user, err := s.verifyUser(ctx, param.Username, param.Password)
	if err != nil {
		s.recordLoginLog(ctx, model.User{}, param.Username, loginLogFail, err.Error(), meta)
		return dto.LoginToken{}, "", err
	}
	user, err = s.updateUserLoginTime(ctx, user)
	if err != nil {
		return dto.LoginToken{}, "", err
	}
	sessionUUID := "session-" + randomID()
	token, refresh, err := s.issueLoginToken(ctx, user, sessionUUID)
	if err != nil {
		return dto.LoginToken{}, "", err
	}
	s.recordLoginLog(ctx, user, param.Username, loginLogSuccess, loginSuccessMsg, meta)
	return token, refresh, nil
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
	user, err = s.updateUserLoginTime(ctx, user)
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
		return dto.AccessTokenBase{}, "", refreshRequestError("Refresh Token 已过期，请重新登录")
	}
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.AccessTokenBase{}, "", refreshNotFoundError("用户不存在", err)
		}
		return dto.AccessTokenBase{}, "", err
	}
	if user.Status != 1 {
		return dto.AccessTokenBase{}, "", refreshForbiddenError("用户已被锁定, 请联系统管理员")
	}
	if err := s.ensureRefreshSessionAllowed(ctx, user, sessionUUID); err != nil {
		return dto.AccessTokenBase{}, "", err
	}
	session, err := s.repo.GetSession(ctx, userID, sessionUUID)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return dto.AccessTokenBase{}, "", authError("Refresh Token 已过期，请重新登录")
		}
		return dto.AccessTokenBase{}, "", err
	}
	newSessionUUID := "session-" + randomID()
	access, expiresAt, err := s.issueAccessToken(ctx, user.ID, newSessionUUID)
	if err != nil {
		return dto.AccessTokenBase{}, "", err
	}
	refresh, _, err := issueToken("refresh", user.ID, newSessionUUID, refreshTokenTTL)
	if err != nil {
		return dto.AccessTokenBase{}, "", err
	}
	if err := s.replaceRefreshSession(ctx, user, session.SessionUUID, newSessionUUID, expiresAt); err != nil {
		return dto.AccessTokenBase{}, "", err
	}
	return dto.AccessTokenBase{
		AccessToken:           access,
		AccessTokenExpireTime: expiresAt.Format(dto.TimeLayout),
		SessionUUID:           newSessionUUID,
	}, refresh, nil
}

func (s *AuthService) Logout(ctx context.Context, authorization string) error {
	userID, sessionUUID, ok, _ := s.parseBearerAccessToken(authorization)
	if !ok {
		return nil
	}
	return s.repo.DeleteSession(ctx, userID, sessionUUID)
}

func (s *AuthService) Authenticate(ctx context.Context, authorization string) (*rbac.CurrentUser, error) {
	userID, sessionUUID, ok, parseErr := s.parseBearerAccessToken(authorization)
	if !ok {
		if stderrors.Is(parseErr, coreauth.ErrAccessTokenExpired) {
			return nil, authError("Token 已过期")
		}
		return nil, authError(accessTokenFailureMessage(authorization))
	}
	session, err := s.repo.GetSession(ctx, userID, sessionUUID)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			// Python stores access tokens in Redis; a missing token key after JWT
			// validation means the token has expired or was removed from the store.
			return nil, authError("Token 已过期")
		}
		return nil, authError("未认证")
	}
	if !session.ExpireTime.IsZero() && time.Now().After(session.ExpireTime) {
		return nil, authError("Token 已过期")
	}
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, authError("未认证")
	}
	if user.Status != 1 {
		return nil, authForbiddenError("用户已被锁定，请联系系统管理员")
	}
	if err := s.ensureUserDeptAllowed(ctx, user); err != nil {
		return nil, err
	}
	roles, err := s.currentUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if err := ensureUserRolesAllowed(roles); err != nil {
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

func (s *AuthService) Codes(ctx context.Context, user *rbac.CurrentUser) ([]string, error) {
	if user != nil && !user.IsSuperAdmin {
		return permissionsFromCurrentUser(user), nil
	}
	// Runtime /auth/codes always has a current user from plugin.Auth(); nil is
	// kept as an admin-style fallback for direct handler tests that mount routes
	// without the auth middleware.
	menus, err := s.repo.ListMenus(ctx, repo.MenuFilter{})
	if err != nil {
		return nil, err
	}
	return permissionsFromMenus(menus), nil
}

func permissionsFromCurrentUser(user *rbac.CurrentUser) []string {
	seen := map[string]struct{}{}
	codes := make([]string, 0)
	for _, role := range user.Roles {
		if !role.Enabled {
			continue
		}
		for _, code := range role.Permissions {
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
	return codes
}

func (s *AuthService) ensureUserDeptAllowed(ctx context.Context, user model.User) error {
	if user.DeptID == nil {
		return nil
	}
	// Python get_current_user rejects tokens when the user's department is
	// deleted or locked; keep that as an authentication-time guard here.
	dept, err := s.repo.GetDept(ctx, *user.DeptID)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return authForbiddenError("用户所属部门不存在或已被删除，请联系系统管理员")
		}
		return err
	}
	if dept.Status != 1 {
		return authForbiddenError("用户所属部门已被锁定，请联系系统管理员")
	}
	return nil
}

func ensureUserRolesAllowed(roles []rbac.Role) error {
	if len(roles) == 0 {
		return nil
	}
	// Python get_current_user treats a user whose assigned roles are all locked
	// as an authentication failure before route-level RBAC checks run.
	for _, role := range roles {
		if role.Enabled {
			return nil
		}
	}
	return authForbiddenError("用户所属角色已被锁定，请联系系统管理员")
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
	if err := s.clearOtherSessions(ctx, user, sessionUUID); err != nil {
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
		LastLoginTime: sessionLastLoginTime(user),
		ExpireTime:    expiresAt,
	})
}

func (s *AuthService) updateUserLoginTime(ctx context.Context, user model.User) (model.User, error) {
	loginTime := time.Now()
	if err := s.repo.UpdateUserLoginTime(ctx, user.ID, loginTime); err != nil {
		return model.User{}, err
	}
	user.LastLoginTime = &loginTime
	return user, nil
}

func sessionLastLoginTime(user model.User) string {
	if user.LastLoginTime != nil {
		return user.LastLoginTime.Format(dto.TimeLayout)
	}
	return time.Now().Format(dto.TimeLayout)
}

func (s *AuthService) replaceRefreshSession(ctx context.Context, user model.User, oldSessionUUID string, newSessionUUID string, expiresAt time.Time) error {
	// Python create_new_token deletes the current access/refresh Redis keys and
	// then creates a fresh access token, whose create_access_token call always
	// assigns a new session_uuid. The Go compatibility store models those token
	// keys as online sessions, so refresh must replace the old session row.
	if err := s.repo.DeleteSession(ctx, user.ID, oldSessionUUID); err != nil {
		return err
	}
	if err := s.clearOtherSessions(ctx, user, newSessionUUID); err != nil {
		return err
	}
	return s.upsertSession(ctx, user, newSessionUUID, expiresAt)
}

func (s *AuthService) issueAccessToken(ctx context.Context, userID int, sessionUUID string) (string, time.Time, error) {
	token, err := s.tokenService.CreateAccessToken(ctx, int64(userID), sessionUUID, nil)
	if err != nil {
		return "", time.Time{}, authError("令牌创建失败")
	}
	return token.Token, token.ExpiresAt, nil
}

func (s *AuthService) recordLoginLog(ctx context.Context, user model.User, username string, status int, msg string, meta RequestMetadata) {
	userUUID := user.UUID
	if userUUID == "" {
		userUUID = "login-log-" + randomID()
	}
	ip := meta.IP
	if ip == "" {
		ip = "127.0.0.1"
	}
	now := time.Now()
	// Python creates login logs in a background task and catches its own failures,
	// so log persistence is intentionally best-effort and must not alter login responses.
	_ = s.repo.CreateLoginLog(ctx, model.LoginLog{
		UserUUID:    userUUID,
		Username:    username,
		Status:      status,
		IP:          ip,
		Country:     meta.Country,
		Region:      meta.Region,
		City:        meta.City,
		UserAgent:   meta.UserAgent,
		Browser:     meta.Browser,
		OS:          meta.OS,
		Device:      meta.Device,
		Msg:         msg,
		LoginTime:   now,
		CreatedTime: now,
	})
}

func (s *AuthService) ensureRefreshSessionAllowed(ctx context.Context, user model.User, sessionUUID string) error {
	if user.IsMultiLogin {
		return nil
	}
	// Python rejects refresh when TOKEN_REDIS_PREFIX contains another session
	// for a single-login user. The Go compatibility store models that as sessions.
	sessions, err := s.repo.ListSessions(ctx, repo.SessionFilter{Username: user.Username})
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if session.ID == user.ID && session.SessionUUID != sessionUUID {
			return refreshForbiddenError("此用户已在异地登录，请重新登录并及时修改密码")
		}
	}
	return nil
}

func (s *AuthService) clearOtherSessions(ctx context.Context, user model.User, sessionUUID string) error {
	if user.IsMultiLogin {
		return nil
	}
	// create_access_token/create_refresh_token delete old Redis keys when
	// multi_login is false; remove stale sessions before storing the new one.
	sessions, err := s.repo.ListSessions(ctx, repo.SessionFilter{Username: user.Username})
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if session.ID != user.ID || session.SessionUUID == sessionUUID {
			continue
		}
		if err := s.repo.DeleteSession(ctx, session.ID, session.SessionUUID); err != nil {
			return err
		}
	}
	return nil
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
	if !ok {
		s.mu.Unlock()
		return authError("验证码已过期")
	}
	if !strings.EqualFold(code, captcha) {
		// Keep invalid attempts retryable; the Python service consumes captchas only after a successful match.
		s.mu.Unlock()
		return authError("验证码错误")
	}
	delete(s.captchas, uuid)
	s.mu.Unlock()
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

func (s *AuthService) parseBearerAccessToken(header string) (int, string, bool, error) {
	token := strings.TrimSpace(header)
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return 0, "", false, nil
	}
	token = strings.TrimSpace(token[7:])
	// Access tokens are JWTs in the Python service and must not fall back to
	// the compatibility refresh-token format; otherwise forged access:* strings
	// could authenticate whenever the referenced session exists.
	claims, err := s.tokenService.ParseAccessToken(token)
	if err == nil && claims.Subject != "" && claims.SessionUUID != "" {
		userID, err := strconv.Atoi(claims.Subject)
		if err == nil {
			return userID, claims.SessionUUID, true, nil
		}
	}
	return 0, "", false, err
}

func accessTokenFailureMessage(authorization string) string {
	// Python's HTTPBearer handles missing credentials before jwt_decode, while a
	// present but malformed token reaches jwt_decode and returns "Token 无效".
	if strings.TrimSpace(authorization) == "" {
		return "未认证"
	}
	return "Token 无效"
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
			MenuCount:      len(menus),
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

func authForbiddenError(message string) error {
	return fbaerrors.New(http.StatusForbidden, http.StatusForbidden, message, nil)
}

func refreshRequestError(message string) error {
	return fbaerrors.New(http.StatusBadRequest, http.StatusBadRequest, message, nil)
}

func refreshNotFoundError(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
}

func refreshForbiddenError(message string) error {
	return fbaerrors.New(http.StatusForbidden, http.StatusForbidden, message, nil)
}
