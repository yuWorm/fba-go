package realtime

import (
	"context"
	"errors"
	"strings"

	coreauth "github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
)

var (
	ErrMissingAuth    = errors.New("realtime auth missing token or session_uuid")
	ErrInvalidAuth    = errors.New("realtime auth invalid")
	ErrNoAuthDisabled = errors.New("realtime no-auth marker is disabled in prod")
)

const defaultNoAuthMarker = "internal"

type AuthPayload struct {
	Token       string `json:"token"`
	SessionUUID string `json:"session_uuid"`
}

type Authenticator interface {
	Authenticate(ctx context.Context, payload AuthPayload) error
}

type JWTAuthenticator struct {
	tokenService coreauth.TokenService
	opts         config.Options
}

func NewJWTAuthenticator(tokenService coreauth.TokenService, opts config.Options) *JWTAuthenticator {
	if tokenService == nil {
		tokenService = coreauth.NewJWTService(opts.Auth)
	}
	return &JWTAuthenticator{tokenService: tokenService, opts: opts.WithDefaults()}
}

func (a *JWTAuthenticator) Authenticate(_ context.Context, payload AuthPayload) error {
	token := strings.TrimSpace(payload.Token)
	sessionUUID := strings.TrimSpace(payload.SessionUUID)
	if token == "" || sessionUUID == "" {
		return ErrMissingAuth
	}

	marker := a.opts.Realtime.NoAuthMarker
	if marker == "" {
		marker = defaultNoAuthMarker
	}
	if token == marker {
		if strings.EqualFold(a.opts.App.Environment, "prod") {
			return ErrNoAuthDisabled
		}
		return nil
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
	return nil
}
