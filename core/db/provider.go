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

// Closer is implemented by providers that own database connection pools.
// Runtimes should close only providers they opened themselves.
type Closer interface {
	Close() error
}
