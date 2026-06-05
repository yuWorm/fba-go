package db_test

import (
	"strings"
	"testing"

	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/db"
)

func TestOpenCreatesSQLiteProvider(t *testing.T) {
	provider, err := db.Open(config.DatabaseOptions{
		Driver:   "sqlite",
		WriteDSN: "file:open_database?mode=memory&cache=shared",
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if provider.Write() == nil || provider.Read() == nil {
		t.Fatal("provider returned nil database handle")
	}
	if got := provider.Write().Dialector.Name(); got != "sqlite" {
		t.Fatalf("dialect = %q, want sqlite", got)
	}
}

func TestOpenRejectsMissingDatabaseConfiguration(t *testing.T) {
	_, err := db.Open(config.DatabaseOptions{})
	if err == nil || !strings.Contains(err.Error(), "database driver is required") {
		t.Fatalf("Open() error = %v, want missing driver", err)
	}
}
