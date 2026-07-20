package realtime

import (
	"context"
	"errors"
	"strconv"
	"strings"

	coreauth "github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
)

var (
	ErrMissingAuth = errors.New("realtime auth missing token or session_uuid")
	ErrInvalidAuth = errors.New("realtime auth invalid")
)

type AuthPayload struct {
	Token       string `json:"token"`
	SessionUUID string `json:"session_uuid"`
}

type Authenticator interface {
	Authenticate(ctx context.Context, payload AuthPayload) error
}

// AccessSessionValidator confirms that a signed JWT still maps to an active,
// non-revoked server-side login session.
type AccessSessionValidator interface {
	ValidateRealtimeSession(ctx context.Context, userID int64, sessionUUID string, accessToken string) error
}

type sessionValidatorResolver interface {
	Resolve(target any) bool
}

type JWTAuthenticatorOption func(*JWTAuthenticator)

func WithAccessSessionValidator(validator AccessSessionValidator) JWTAuthenticatorOption {
	return func(authenticator *JWTAuthenticator) {
		authenticator.sessionValidator = validator
	}
}

func WithAccessSessionValidatorResolver(resolver sessionValidatorResolver) JWTAuthenticatorOption {
	return func(authenticator *JWTAuthenticator) {
		authenticator.sessionValidatorResolver = resolver
	}
}

type JWTAuthenticator struct {
	tokenService             coreauth.TokenService
	sessionValidator         AccessSessionValidator
	sessionValidatorResolver sessionValidatorResolver
}

func NewJWTAuthenticator(tokenService coreauth.TokenService, opts config.Options, options ...JWTAuthenticatorOption) *JWTAuthenticator {
	if tokenService == nil {
		tokenService = coreauth.NewJWTService(opts.Auth)
	}
	authenticator := &JWTAuthenticator{tokenService: tokenService}
	for _, option := range options {
		option(authenticator)
	}
	return authenticator
}

func (a *JWTAuthenticator) Authenticate(ctx context.Context, payload AuthPayload) error {
	token := strings.TrimSpace(payload.Token)
	sessionUUID := strings.TrimSpace(payload.SessionUUID)
	if token == "" || sessionUUID == "" {
		return ErrMissingAuth
	}

	token = strings.TrimPrefix(token, "Bearer ")
	claims, err := a.tokenService.ParseAccessToken(token)
	if err != nil {
		return ErrInvalidAuth
	}
	// Python receives both token and session_uuid in the Socket.IO auth payload.
	// Binding them here prevents a client from reusing a valid JWT while spoofing
	// another online session identifier.
	if claims.SessionUUID != sessionUUID {
		return ErrInvalidAuth
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return ErrInvalidAuth
	}
	validator := a.sessionValidator
	if validator == nil && a.sessionValidatorResolver != nil {
		_ = a.sessionValidatorResolver.Resolve(&validator)
	}
	if validator == nil {
		return ErrInvalidAuth
	}
	if err := validator.ValidateRealtimeSession(ctx, userID, sessionUUID, token); err != nil {
		return ErrInvalidAuth
	}
	return nil
}
