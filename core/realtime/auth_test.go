package realtime_test

import (
	"context"
	"errors"
	"testing"

	coreauth "github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/realtime"
)

func TestJWTAuthenticatorAcceptsMatchingTokenSession(t *testing.T) {
	tokenService := coreauth.NewJWTService(config.AuthOptions{JWTSecret: "secret"})
	token, err := tokenService.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	authenticator := realtime.NewJWTAuthenticator(tokenService, config.Options{Auth: config.AuthOptions{JWTSecret: "secret"}})

	if err := authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: token.Token, SessionUUID: "session-1"}); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
}

func TestJWTAuthenticatorRejectsMismatchedSession(t *testing.T) {
	tokenService := coreauth.NewJWTService(config.AuthOptions{JWTSecret: "secret"})
	token, err := tokenService.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	authenticator := realtime.NewJWTAuthenticator(tokenService, config.Options{Auth: config.AuthOptions{JWTSecret: "secret"}})

	err = authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: token.Token, SessionUUID: "session-2"})
	if !errors.Is(err, realtime.ErrInvalidAuth) {
		t.Fatalf("Authenticate() error = %v, want ErrInvalidAuth", err)
	}
}

func TestJWTAuthenticatorAllowsNoAuthMarkerOutsideProd(t *testing.T) {
	authenticator := realtime.NewJWTAuthenticator(nil, config.Options{
		App:      config.AppOptions{Environment: "dev"},
		Realtime: config.RealtimeOptions{NoAuthMarker: "internal"},
	})

	if err := authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: "internal", SessionUUID: "session-1"}); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
}

func TestJWTAuthenticatorRejectsNoAuthMarkerInProd(t *testing.T) {
	authenticator := realtime.NewJWTAuthenticator(nil, config.Options{
		App:      config.AppOptions{Environment: "prod"},
		Realtime: config.RealtimeOptions{NoAuthMarker: "internal"},
	})

	err := authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: "internal", SessionUUID: "session-1"})
	if !errors.Is(err, realtime.ErrNoAuthDisabled) {
		t.Fatalf("Authenticate() error = %v, want ErrNoAuthDisabled", err)
	}
}
