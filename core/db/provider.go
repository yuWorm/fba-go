package db

import (
	"context"

	"gorm.io/gorm"
)

type Provider interface {
	Write() *gorm.DB
	Read() *gorm.DB
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}
