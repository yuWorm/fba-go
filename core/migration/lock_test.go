package migration_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/yuWorm/fba-go/core/migration"
	"gorm.io/gorm"
)

func TestGORMAdvisoryLockSerializesSQLiteRuntimes(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "migration.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	first := migration.NewGORMAdvisoryLock(database, "test:migrations")
	second := migration.NewGORMAdvisoryLock(database, "test:migrations")
	releaseFirst, err := first.Acquire(context.Background())
	if err != nil {
		t.Fatalf("first Acquire() error = %v", err)
	}

	blockedCtx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	if _, err := second.Acquire(blockedCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second Acquire() error = %v, want deadline exceeded", err)
	}
	if err := releaseFirst(context.Background()); err != nil {
		t.Fatalf("first release error = %v", err)
	}

	releaseSecond, err := second.Acquire(context.Background())
	if err != nil {
		t.Fatalf("second Acquire() after release error = %v", err)
	}
	if err := releaseSecond(context.Background()); err != nil {
		t.Fatalf("second release error = %v", err)
	}
}
