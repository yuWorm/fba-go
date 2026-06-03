package repo

import (
	"context"
	"errors"

	"github.com/yuWorm/fba-plugin-oauth2/model"
)

var ErrNotFound = errors.New("not found")

type CreateUserSocialParam struct {
	SID    string
	Source string
	UserID int
}

type Repository interface {
	GetBySID(ctx context.Context, sid string, source string) (model.UserSocial, error)
	CheckBinding(ctx context.Context, userID int, source string) (model.UserSocial, error)
	ListByUserID(ctx context.Context, userID int) ([]model.UserSocial, error)
	Create(ctx context.Context, param CreateUserSocialParam) (model.UserSocial, error)
	Delete(ctx context.Context, userID int, source string) (int, error)
}

func SeedData() []model.UserSocial {
	return nil
}
