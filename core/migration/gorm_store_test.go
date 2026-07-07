package migration_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/yuWorm/fba-go/core/migration"
	"gorm.io/gorm"
)

func TestGORMStoreRecordsAndReportsAppliedMigration(t *testing.T) {
	database := openMigrationStoreDB(t)
	store := migration.NewGORMStore(database)
	ctx := context.Background()

	applied, err := store.IsApplied(ctx, "plugin:admin", "0001")
	if err != nil {
		t.Fatalf("IsApplied() error = %v", err)
	}
	if applied {
		t.Fatal("IsApplied() = true before record, want false")
	}

	if err := store.Record(ctx, migration.Record{
		Scope:   "plugin:admin",
		Version: "0001",
		Name:    "auto migrate",
		Success: true,
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	applied, err = store.IsApplied(ctx, "plugin:admin", "0001")
	if err != nil {
		t.Fatalf("IsApplied() error = %v", err)
	}
	if !applied {
		t.Fatal("IsApplied() = false after successful record, want true")
	}
}

func TestGORMStoreUpdatesFailedMigrationRecordOnRetry(t *testing.T) {
	database := openMigrationStoreDB(t)
	store := migration.NewGORMStore(database)
	ctx := context.Background()

	if err := store.Record(ctx, migration.Record{
		Scope:   "plugin:admin",
		Version: "0001",
		Success: false,
		Error:   "boom",
	}); err != nil {
		t.Fatalf("Record(failed) error = %v", err)
	}
	applied, err := store.IsApplied(ctx, "plugin:admin", "0001")
	if err != nil {
		t.Fatalf("IsApplied() error = %v", err)
	}
	if applied {
		t.Fatal("failed migration was reported as applied")
	}

	if err := store.Record(ctx, migration.Record{
		Scope:   "plugin:admin",
		Version: "0001",
		Success: true,
	}); err != nil {
		t.Fatalf("Record(success retry) error = %v", err)
	}
	var count int64
	if err := database.Table("fba_migration_records").Count(&count).Error; err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("record count = %d, want 1", count)
	}
}

func openMigrationStoreDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:migration_store?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	return database
}
