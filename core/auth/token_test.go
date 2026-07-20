package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/redisx"
)

func TestJWTServiceCreatesCompatiblePayload(t *testing.T) {
	service := auth.NewJWTService(config.AuthOptions{
		JWTSecret:      "0123456789abcdef0123456789abcdef",
		JWTIssuer:      "token-test",
		AccessTokenTTL: time.Hour,
	})
	now := time.Now().UTC().Truncate(time.Second)
	service.Now = func() time.Time {
		return now
	}

	token, err := service.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	claims, err := service.ParseAccessToken(token.Token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	if claims.SessionUUID != "session-1" {
		t.Fatalf("SessionUUID = %q, want session-1", claims.SessionUUID)
	}
	if claims.Subject != "10001" {
		t.Fatalf("Subject = %q, want string user id", claims.Subject)
	}
	wantExp := now.Add(time.Hour).Unix()
	if claims.ExpiresAt.Unix() != wantExp {
		t.Fatalf("ExpiresAt = %d, want %d", claims.ExpiresAt.Unix(), wantExp)
	}
}

func TestJWTServiceCreatesUniqueTokensForSameUserSession(t *testing.T) {
	service := auth.NewJWTService(config.AuthOptions{
		JWTSecret:      "0123456789abcdef0123456789abcdef",
		JWTIssuer:      "token-test",
		AccessTokenTTL: time.Hour,
	})
	now := time.Now().UTC().Truncate(time.Second)
	service.Now = func() time.Time {
		return now
	}

	first, err := service.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken(first) error = %v", err)
	}
	second, err := service.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken(second) error = %v", err)
	}

	if first.Token == second.Token {
		t.Fatal("tokens are equal, want unique token per issuance")
	}
}

func TestJWTServiceReportsExpiredAccessToken(t *testing.T) {
	service := auth.NewJWTService(config.AuthOptions{
		JWTSecret:      "0123456789abcdef0123456789abcdef",
		JWTIssuer:      "token-test",
		AccessTokenTTL: time.Hour,
	})
	issuedAt := time.Now().UTC().Truncate(time.Second).Add(-2 * time.Hour)
	service.Now = func() time.Time {
		return issuedAt
	}
	token, err := service.CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	_, err = service.ParseAccessToken(token.Token)
	if !errors.Is(err, auth.ErrAccessTokenExpired) {
		t.Fatalf("ParseAccessToken() error = %v, want ErrAccessTokenExpired", err)
	}
}

func TestValidateJWTOptionsRejectsUnsafeConfiguration(t *testing.T) {
	tests := []struct {
		name string
		opts config.AuthOptions
	}{
		{name: "missing secret", opts: config.AuthOptions{JWTIssuer: "issuer"}},
		{name: "short secret", opts: config.AuthOptions{JWTSecret: "too-short", JWTIssuer: "issuer"}},
		{name: "missing issuer", opts: config.AuthOptions{JWTSecret: "0123456789abcdef0123456789abcdef"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := auth.ValidateJWTOptions(test.opts); err == nil {
				t.Fatal("ValidateJWTOptions() error = nil")
			}
		})
	}
}

func TestJWTServiceRejectsTokenFromDifferentIssuer(t *testing.T) {
	issuerA := config.AuthOptions{
		JWTSecret: "0123456789abcdef0123456789abcdef",
		JWTIssuer: "issuer-a",
	}
	issuerB := issuerA
	issuerB.JWTIssuer = "issuer-b"
	token, err := auth.NewJWTService(issuerA).CreateAccessToken(context.Background(), 10001, "session-1", nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	if _, err := auth.NewJWTService(issuerB).ParseAccessToken(token.Token); err == nil {
		t.Fatal("ParseAccessToken() error = nil for a different issuer")
	}
}

func TestSessionKeysUseCompatibleRedisKeys(t *testing.T) {
	keys := auth.NewSessionKeys(redisx.NewKeys(""))
	got := keys.ForSession(10001, "session-1")

	if got.AccessToken != "fba:token:10001:session-1" {
		t.Fatalf("AccessToken key = %q", got.AccessToken)
	}
	if got.RefreshToken != "fba:refresh_token:10001:session-1" {
		t.Fatalf("RefreshToken key = %q", got.RefreshToken)
	}
	if got.UserCache != "fba:user:10001" {
		t.Fatalf("UserCache key = %q", got.UserCache)
	}
}
