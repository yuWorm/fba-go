package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/redisx"
)

func TestJWTServiceCreatesCompatiblePayload(t *testing.T) {
	service := auth.NewJWTService(config.AuthOptions{
		JWTSecret:      "secret",
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
