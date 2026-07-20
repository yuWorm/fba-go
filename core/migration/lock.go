package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"

	"gorm.io/gorm"
)

type Lock interface {
	Acquire(ctx context.Context) (release func(context.Context) error, err error)
}

type NoopLock struct{}

func (NoopLock) Acquire(context.Context) (func(context.Context) error, error) {
	return func(context.Context) error { return nil }, nil
}

type GORMAdvisoryLock struct {
	database *gorm.DB
	name     string
}

func NewGORMAdvisoryLock(database *gorm.DB, name string) *GORMAdvisoryLock {
	if name == "" {
		name = "fba-go:migrations"
	}
	return &GORMAdvisoryLock{database: database, name: name}
}

func (l *GORMAdvisoryLock) Acquire(ctx context.Context) (func(context.Context) error, error) {
	if l == nil || l.database == nil {
		return nil, gorm.ErrInvalidDB
	}
	database, err := l.database.DB()
	if err != nil {
		return nil, err
	}
	switch l.database.Dialector.Name() {
	case "postgres":
		return acquirePostgresLock(ctx, database, l.name)
	case "mysql":
		return acquireMySQLLock(ctx, database, l.name)
	case "sqlite":
		// SQLite serializes file writes itself. This process lock closes the
		// check-before-migrate race between concurrent runtimes sharing a pool.
		return acquireLocalLock(ctx, fmt.Sprintf("%p:%s", database, l.name))
	default:
		return nil, fmt.Errorf("migration advisory lock does not support database driver %q", l.database.Dialector.Name())
	}
}

func acquirePostgresLock(ctx context.Context, database *sql.DB, name string) (func(context.Context) error, error) {
	connection, err := database.Conn(ctx)
	if err != nil {
		return nil, err
	}
	key := advisoryLockKey(name)
	if _, err := connection.ExecContext(ctx, "SELECT pg_advisory_lock($1)", key); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return func(releaseCtx context.Context) error {
		_, unlockErr := connection.ExecContext(releaseCtx, "SELECT pg_advisory_unlock($1)", key)
		return errors.Join(unlockErr, connection.Close())
	}, nil
}

func acquireMySQLLock(ctx context.Context, database *sql.DB, name string) (func(context.Context) error, error) {
	connection, err := database.Conn(ctx)
	if err != nil {
		return nil, err
	}
	lockName := fmt.Sprintf("fba-go:%x", advisoryLockKey(name))
	for {
		var acquired sql.NullInt64
		if err := connection.QueryRowContext(ctx, "SELECT GET_LOCK(?, 1)", lockName).Scan(&acquired); err != nil {
			_ = connection.Close()
			return nil, err
		}
		if acquired.Valid && acquired.Int64 == 1 {
			break
		}
		if !acquired.Valid {
			_ = connection.Close()
			return nil, fmt.Errorf("mysql GET_LOCK returned NULL")
		}
		if err := ctx.Err(); err != nil {
			_ = connection.Close()
			return nil, err
		}
	}
	return func(releaseCtx context.Context) error {
		var released sql.NullInt64
		unlockErr := connection.QueryRowContext(releaseCtx, "SELECT RELEASE_LOCK(?)", lockName).Scan(&released)
		if unlockErr == nil && (!released.Valid || released.Int64 != 1) {
			unlockErr = fmt.Errorf("mysql RELEASE_LOCK did not release %q", lockName)
		}
		return errors.Join(unlockErr, connection.Close())
	}, nil
}

var localLocks sync.Map

func acquireLocalLock(ctx context.Context, name string) (func(context.Context) error, error) {
	candidate := make(chan struct{}, 1)
	value, _ := localLocks.LoadOrStore(name, candidate)
	lock := value.(chan struct{})
	select {
	case lock <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	var once sync.Once
	return func(context.Context) error {
		once.Do(func() {
			<-lock
		})
		return nil
	}, nil
}

func advisoryLockKey(name string) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(name))
	return int64(hash.Sum64())
}
