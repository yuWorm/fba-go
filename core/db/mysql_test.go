package db

import (
	"strings"
	"testing"
)

func TestNormalizeMySQLDSNForcesUTF8MB4(t *testing.T) {
	got := normalizeMySQLDSN("root:pass@tcp(localhost:3306)/fbago?charset=utf8mb3&parseTime=True&loc=Local")
	if strings.Contains(got, "charset=utf8mb3") {
		t.Fatalf("dsn still contains utf8mb3: %s", got)
	}
	if !strings.Contains(got, "charset=utf8mb4") {
		t.Fatalf("dsn = %s, want charset=utf8mb4", got)
	}
}

func TestMySQLTableOptionsUseUTF8MB4(t *testing.T) {
	options := mysqlTableOptions()
	if !strings.Contains(options, "CHARSET=utf8mb4") {
		t.Fatalf("table options = %q, want utf8mb4 charset", options)
	}
	if !strings.Contains(options, "COLLATE=utf8mb4_unicode_ci") {
		t.Fatalf("table options = %q, want utf8mb4 collation", options)
	}
}
