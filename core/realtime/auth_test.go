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
	tokenService := coreauth.NewJWTService(validAuthOptions())
	token, err := tokenService.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	authenticator := realtime.NewJWTAuthenticator(tokenService, config.Options{Auth: validAuthOptions()}, realtime.WithAccessSessionValidator(
		accessSessionValidatorFunc(func(_ context.Context, userID int64, sessionUUID string, accessToken string) error {
			if userID != 10001 || sessionUUID != "session-1" || accessToken != token.Token {
				t.Fatalf("validator arguments = (%d, %q, %q)", userID, sessionUUID, accessToken)
			}
			return nil
		}),
	))

	if err := authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: token.Token, SessionUUID: "session-1"}); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
}

func TestJWTAuthenticatorRejectsMismatchedSession(t *testing.T) {
	tokenService := coreauth.NewJWTService(validAuthOptions())
	token, err := tokenService.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	authenticator := realtime.NewJWTAuthenticator(tokenService, config.Options{Auth: validAuthOptions()})

	err = authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: token.Token, SessionUUID: "session-2"})
	if !errors.Is(err, realtime.ErrInvalidAuth) {
		t.Fatalf("Authenticate() error = %v, want ErrInvalidAuth", err)
	}
}

func TestJWTAuthenticatorRejectsValidJWTWithoutActiveSessionValidator(t *testing.T) {
	tokenService := coreauth.NewJWTService(validAuthOptions())
	token, err := tokenService.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	authenticator := realtime.NewJWTAuthenticator(tokenService, config.Options{Auth: validAuthOptions()})

	err = authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: token.Token, SessionUUID: "session-1"})
	if !errors.Is(err, realtime.ErrInvalidAuth) {
		t.Fatalf("Authenticate() error = %v, want ErrInvalidAuth", err)
	}
}

func TestJWTAuthenticatorRejectsRevokedSession(t *testing.T) {
	tokenService := coreauth.NewJWTService(validAuthOptions())
	token, err := tokenService.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	authenticator := realtime.NewJWTAuthenticator(tokenService, config.Options{Auth: validAuthOptions()}, realtime.WithAccessSessionValidator(
		accessSessionValidatorFunc(func(context.Context, int64, string, string) error {
			return errors.New("session revoked")
		}),
	))

	err = authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: token.Token, SessionUUID: "session-1"})
	if !errors.Is(err, realtime.ErrInvalidAuth) {
		t.Fatalf("Authenticate() error = %v, want ErrInvalidAuth", err)
	}
}

func TestJWTAuthenticatorRejectsFormerNoAuthMarkerInEveryEnvironment(t *testing.T) {
	authenticator := realtime.NewJWTAuthenticator(coreauth.NewJWTService(validAuthOptions()), config.Options{Auth: validAuthOptions()})

	err := authenticator.Authenticate(context.Background(), realtime.AuthPayload{Token: "internal", SessionUUID: "session-1"})
	if !errors.Is(err, realtime.ErrInvalidAuth) {
		t.Fatalf("Authenticate() error = %v, want ErrInvalidAuth", err)
	}
}

type accessSessionValidatorFunc func(context.Context, int64, string, string) error

func (f accessSessionValidatorFunc) ValidateRealtimeSession(ctx context.Context, userID int64, sessionUUID string, accessToken string) error {
	return f(ctx, userID, sessionUUID, accessToken)
}

func validAuthOptions() config.AuthOptions {
	return config.AuthOptions{
		JWTSecret: "0123456789abcdef0123456789abcdef",
		JWTIssuer: "realtime-test",
	}
}
