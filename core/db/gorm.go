package db

import (
	"context"

	"gorm.io/gorm"
)

type GORMProvider struct {
	write *gorm.DB
	read  *gorm.DB
}

func NewGORMProvider(write *gorm.DB, read *gorm.DB) *GORMProvider {
	if read == nil {
		read = write
	}
	return &GORMProvider{write: write, read: read}
}

func (p *GORMProvider) Write() *gorm.DB {
	return p.write
}

func (p *GORMProvider) Read() *gorm.DB {
	return p.read
}

func (p *GORMProvider) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return p.write.WithContext(ctx).Transaction(fn)
}
