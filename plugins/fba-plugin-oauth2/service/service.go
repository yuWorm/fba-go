package service

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	admindto "github.com/yuWorm/fba-plugin-admin/dto"
	adminmodel "github.com/yuWorm/fba-plugin-admin/model"
	adminrepo "github.com/yuWorm/fba-plugin-admin/repo"
	adminservice "github.com/yuWorm/fba-plugin-admin/service"
	"github.com/yuWorm/fba-plugin-oauth2/repo"
)

var (
	ErrStateNotFound         = errors.New("oauth2 state not found")
	ErrProviderNotConfigured = errors.New("oauth2 provider is not configured")
	usernameUnsupportedChars = regexp.MustCompile(`[^A-Za-z0-9_-]+`)
	// The Python OAuth2 DAO creates users without an explicit department; the
	// Go admin repository requires a valid dept id, so use the seeded root dept.
	defaultOAuth2UserDeptID = 1
)

type Service struct {
	repo       repo.Repository
	adminRepo  adminrepo.Repository
	users      *adminservice.UserService
	auth       *adminservice.AuthService
	stateStore StateStore
	providers  map[string]Provider
	settings   Settings
}

type Options struct {
	Repository repo.Repository
	AdminRepo  adminrepo.Repository
	Users      *adminservice.UserService
	Auth       *adminservice.AuthService
	StateStore StateStore
	Providers  map[string]Provider
	Settings   Settings
}

type CallbackResult struct {
	Binding     bool
	AccessToken string
	SessionUUID string
	Refresh     string
}

func New(options Options) *Service {
	settings := options.Settings.withDefaults()
	repository := options.Repository
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	adminRepository := options.AdminRepo
	if adminRepository == nil {
		adminRepository = adminrepo.NewMemoryRepository(adminrepo.SeedData())
	}
	stateStore := options.StateStore
	if stateStore == nil {
		stateStore = NewMemoryStateStore()
	}
	providers := options.Providers
	if providers == nil {
		providers = DefaultProviders(settings)
	}
	users := options.Users
	if users == nil {
		users = adminservice.NewUserService(adminRepository)
	}
	auth := options.Auth
	if auth == nil {
		auth = adminservice.NewAuthService(adminRepository)
	}
	return &Service{
		repo:       repository,
		adminRepo:  adminRepository,
		users:      users,
		auth:       auth,
		stateStore: stateStore,
		providers:  providers,
		settings:   settings,
	}
}

func (s *Service) Settings() Settings {
	return s.settings
}

func (s *Service) LoginAuthURL(ctx context.Context, source string) (string, error) {
	normalized, err := normalizeSource(source)
	if err != nil {
		return "", err
	}
	state := uuid.NewString()
	if err := s.stateStore.Set(ctx, state, StatePayload{Type: AuthTypeLogin}, s.settings.StateExpire); err != nil {
		return "", err
	}
	provider, err := s.provider(normalized)
	if err != nil {
		return "", err
	}
	return provider.AuthorizationURL(state, s.redirectURI(normalized))
}

func (s *Service) BindingAuthURL(ctx context.Context, userID int, source string) (string, error) {
	normalized, err := normalizeSource(source)
	if err != nil {
		return "", err
	}
	state := uuid.NewString()
	if err := s.stateStore.Set(ctx, state, StatePayload{Type: AuthTypeBinding, UserID: userID}, s.settings.StateExpire); err != nil {
		return "", err
	}
	provider, err := s.provider(normalized)
	if err != nil {
		return "", err
	}
	return provider.AuthorizationURL(state, s.redirectURI(normalized))
}

func (s *Service) Bindings(ctx context.Context, userID int) ([]string, error) {
	bindings, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(bindings))
	for _, item := range bindings {
		result = append(result, item.Source)
	}
	return result, nil
}

func (s *Service) Unbind(ctx context.Context, userID int, source string) error {
	normalized, err := normalizeSource(source)
	if err != nil {
		return err
	}
	if _, err := s.repo.CheckBinding(ctx, userID, normalized); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return notFoundError("用户未绑定 "+normalized+" 账号", err)
		}
		return err
	}
	if _, err := s.repo.Delete(ctx, userID, normalized); err != nil {
		return err
	}
	return nil
}

func (s *Service) Callback(ctx context.Context, source string, code string, state string, meta adminservice.RequestMetadata) (CallbackResult, error) {
	normalized, err := normalizeSource(source)
	if err != nil {
		return CallbackResult{}, err
	}
	if strings.TrimSpace(code) == "" {
		return CallbackResult{}, forbiddenError("OAuth2 授权码缺失", nil)
	}
	provider, err := s.provider(normalized)
	if err != nil {
		return CallbackResult{}, err
	}
	providerUser, err := provider.UserInfo(ctx, code, s.redirectURI(normalized))
	if err != nil {
		if errors.Is(err, ErrProviderNotConfigured) {
			return CallbackResult{}, badRequestError("OAuth2 provider credentials are not configured", err)
		}
		return CallbackResult{}, err
	}
	account, err := oauth2AccountFromProvider(normalized, providerUser)
	if err != nil {
		return CallbackResult{}, err
	}
	payload, err := s.stateStore.Pop(ctx, state)
	if err != nil {
		if errors.Is(err, ErrStateNotFound) {
			return CallbackResult{}, forbiddenError("OAuth2 状态信息无效或缺失", err)
		}
		return CallbackResult{}, err
	}
	if payload.Type == AuthTypeBinding {
		if payload.UserID <= 0 {
			return CallbackResult{}, forbiddenError("非法操作，OAuth2 状态信息无效", nil)
		}
		if err := s.BindingWithOAuth2(ctx, payload.UserID, account); err != nil {
			return CallbackResult{}, err
		}
		return CallbackResult{Binding: true}, nil
	}
	if payload.Type != AuthTypeLogin {
		return CallbackResult{}, forbiddenError("OAuth2 状态信息无效", nil)
	}
	token, refresh, err := s.Login(ctx, account, meta)
	if err != nil {
		return CallbackResult{}, err
	}
	return CallbackResult{
		AccessToken: token.AccessToken,
		SessionUUID: token.SessionUUID,
		Refresh:     refresh,
	}, nil
}

type OAuth2Account struct {
	SID       string
	Source    string
	Username  string
	Nickname  string
	Email     *string
	AvatarURL *string
}

func (s *Service) Login(ctx context.Context, account OAuth2Account, meta adminservice.RequestMetadata) (admindto.LoginToken, string, error) {
	if account.SID == "" {
		return admindto.LoginToken{}, "", forbiddenError("OAuth2 用户信息缺失", nil)
	}
	social, err := s.repo.GetBySID(ctx, account.SID, account.Source)
	if err == nil {
		user, err := s.adminRepo.GetUser(ctx, social.UserID)
		if err != nil {
			if errors.Is(err, adminrepo.ErrNotFound) {
				return admindto.LoginToken{}, "", notFoundError("用户不存在", err)
			}
			return admindto.LoginToken{}, "", err
		}
		user, err = s.updateMissingAvatar(ctx, user, account.AvatarURL)
		if err != nil {
			return admindto.LoginToken{}, "", err
		}
		return s.auth.OAuth2Login(ctx, user, meta)
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return admindto.LoginToken{}, "", err
	}

	user, err := s.findOrCreateUser(ctx, account)
	if err != nil {
		return admindto.LoginToken{}, "", err
	}
	if _, err := s.repo.Create(ctx, repo.CreateUserSocialParam{
		SID:    account.SID,
		Source: account.Source,
		UserID: user.ID,
	}); err != nil {
		return admindto.LoginToken{}, "", err
	}
	return s.auth.OAuth2Login(ctx, user, meta)
}

func (s *Service) BindingWithOAuth2(ctx context.Context, userID int, account OAuth2Account) error {
	if _, err := s.repo.CheckBinding(ctx, userID, account.Source); err == nil {
		return badRequestError("用户已绑定 "+account.Source+" 账号", nil)
	} else if !errors.Is(err, repo.ErrNotFound) {
		return err
	}
	if _, err := s.repo.GetBySID(ctx, account.SID, account.Source); err == nil {
		return badRequestError("该 "+account.Source+" 账号已被其他用户绑定", nil)
	} else if !errors.Is(err, repo.ErrNotFound) {
		return err
	}
	_, err := s.repo.Create(ctx, repo.CreateUserSocialParam{
		SID:    account.SID,
		Source: account.Source,
		UserID: userID,
	})
	return err
}

func (s *Service) findOrCreateUser(ctx context.Context, account OAuth2Account) (adminmodel.User, error) {
	if account.Email != nil && *account.Email != "" {
		user, err := s.adminRepo.GetUserByEmail(ctx, *account.Email)
		if err == nil {
			return s.updateMissingAvatar(ctx, user, account.AvatarURL)
		}
		if !errors.Is(err, adminrepo.ErrNotFound) {
			return adminmodel.User{}, err
		}
	}
	username, err := s.uniqueUsername(ctx, account.Username)
	if err != nil {
		return adminmodel.User{}, err
	}
	nickname := optionalString(account.Nickname)
	if nickname == nil {
		nickname = optionalString(username)
	}
	// Python stores password=None for OAuth2-created users. The Go admin model
	// keeps password as a string, so use an unguessable placeholder while relying
	// on OAuth2 sessions for authentication.
	detail, err := s.users.Create(ctx, admindto.UserCreateParam{
		Username: username,
		Password: "oauth2-" + uuid.NewString(),
		Nickname: nickname,
		Email:    account.Email,
		DeptID:   defaultOAuth2UserDeptID,
		Roles:    nil,
	})
	if err != nil {
		return adminmodel.User{}, err
	}
	user, err := s.adminRepo.GetUser(ctx, detail.ID)
	if err != nil {
		return adminmodel.User{}, err
	}
	return s.updateMissingAvatar(ctx, user, account.AvatarURL)
}

func (s *Service) updateMissingAvatar(ctx context.Context, user adminmodel.User, avatar *string) (adminmodel.User, error) {
	if user.Avatar != nil || avatar == nil || *avatar == "" {
		return user, nil
	}
	if err := s.adminRepo.UpdateUserAvatar(ctx, user.ID, avatar); err != nil {
		return adminmodel.User{}, err
	}
	user.Avatar = avatar
	return user, nil
}

func (s *Service) uniqueUsername(ctx context.Context, preferred string) (string, error) {
	base := sanitizeUsername(preferred)
	if base == "" {
		base = "oauth2_user"
	}
	candidates := []string{base}
	for i := 0; i < 10; i++ {
		candidates = append(candidates, base+"_"+strings.ReplaceAll(uuid.NewString()[:8], "-", ""))
	}
	for _, candidate := range candidates {
		if _, err := s.adminRepo.GetUserByUsername(ctx, candidate); errors.Is(err, adminrepo.ErrNotFound) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
	}
	return "", conflictError("用户名已存在，请重试", nil)
}

func sanitizeUsername(value string) string {
	value = strings.TrimSpace(value)
	value = usernameUnsupportedChars.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_-")
	if len(value) > 48 {
		value = value[:48]
	}
	return value
}

func oauth2AccountFromProvider(source string, user ProviderUser) (OAuth2Account, error) {
	if user.ID == "" {
		return OAuth2Account{}, forbiddenError("OAuth2 用户信息缺失", nil)
	}
	return OAuth2Account{
		SID:       user.ID,
		Source:    source,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}, nil
}

func (s *Service) provider(source string) (Provider, error) {
	provider := s.providers[source]
	if provider == nil {
		return nil, forbiddenError("暂不支持 "+source+" OAuth2 登录", nil)
	}
	return provider, nil
}

func (s *Service) redirectURI(source string) string {
	if source == SourceGoogle {
		return s.settings.GoogleRedirectURI
	}
	return s.settings.GithubRedirectURI
}

func normalizeSource(source string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "github":
		return SourceGithub, nil
	case "google":
		return SourceGoogle, nil
	default:
		return "", forbiddenError("暂不支持 "+source+" OAuth2 登录", nil)
	}
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func badRequestError(message string, cause error) error {
	return fbaerrors.New(http.StatusBadRequest, http.StatusBadRequest, message, cause)
}

func conflictError(message string, cause error) error {
	return fbaerrors.New(http.StatusConflict, http.StatusConflict, message, cause)
}

func forbiddenError(message string, cause error) error {
	return fbaerrors.New(http.StatusForbidden, http.StatusForbidden, message, cause)
}

func notFoundError(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
}
