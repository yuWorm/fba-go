package service_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	coreauth "github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
	"github.com/yuWorm/fba-plugin-admin/service"
)

func TestAuthenticateReturnsTokenInvalidWhenSessionUserIsMissing(t *testing.T) {
	ctx := context.Background()
	sessionUUID := "orphan-session"
	tokenService := coreauth.NewJWTService(config.AuthOptions{AccessTokenTTL: time.Hour})
	token, err := tokenService.CreateAccessToken(ctx, 999, sessionUUID, nil)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}
	repository := repo.NewMemoryRepository(model.Seed{
		Sessions: []model.Session{{
			ID:          999,
			SessionUUID: sessionUUID,
			Username:    "missing",
			Status:      1,
			ExpireTime:  time.Now().Add(time.Hour),
		}},
	})
	authService := service.NewAuthService(repository)

	_, err = authService.Authenticate(ctx, "Bearer "+token.Token)

	var appErr *fbaerrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Authenticate() error = %T, want AppError", err)
	}
	if appErr.HTTPStatus() != http.StatusUnauthorized || appErr.PublicMessage() != "Token 无效" {
		t.Fatalf("Authenticate() error = (%d, %q), want (401, Token 无效)", appErr.HTTPStatus(), appErr.PublicMessage())
	}
}
