package db

import (
	"context"
	"errors"

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

func (p *GORMProvider) Close() error {
	var closeErrs []error
	var writePool, readPool interface{ Close() error }
	if p.write != nil {
		pool, err := p.write.DB()
		if err != nil {
			closeErrs = append(closeErrs, err)
		} else {
			writePool = pool
		}
	}
	if p.read != nil {
		pool, err := p.read.DB()
		if err != nil {
			closeErrs = append(closeErrs, err)
		} else if pool != writePool {
			readPool = pool
		}
	}
	if readPool != nil {
		closeErrs = append(closeErrs, readPool.Close())
	}
	if writePool != nil {
		closeErrs = append(closeErrs, writePool.Close())
	}
	return errors.Join(closeErrs...)
}

var _ Closer = (*GORMProvider)(nil)
