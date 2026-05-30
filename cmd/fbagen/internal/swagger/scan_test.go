package swagger_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	fbswagger "github.com/yuWorm/fba-go/cmd/fbagen/internal/swagger"
)

func TestScanReadsPluginLockAndWritesAggregatedOpenAPI(t *testing.T) {
	dir := t.TempDir()
	adminFragment := filepath.Join(dir, "admin.swagger.json")
	dictFragment := filepath.Join(dir, "dict.swagger.json")
	lockPath := filepath.Join(dir, "plugin_manifest.lock")
	outPath := filepath.Join(dir, "openapi.json")

	writeJSON(t, adminFragment, map[string]any{
		"plugin_id": "admin",
		"paths": map[string]any{
			"/api/v1/sys/users": map[string]any{
				"get": map[string]any{"summary": "list users"},
			},
		},
		"schemas": map[string]any{
			"User": map[string]any{"type": "object"},
		},
	})
	writeJSON(t, dictFragment, map[string]any{
		"plugin_id": "dict",
		"paths": map[string]any{
			"/api/v1/dict-datas": map[string]any{
				"get": map[string]any{"summary": "list dict data"},
			},
		},
	})
	writeJSON(t, lockPath, map[string]any{
		"plugins": []map[string]any{
			{"id": "admin", "swagger": adminFragment},
			{"id": "dict", "swagger": dictFragment},
		},
	})

	err := fbswagger.Scan(fbswagger.ScanOptions{
		PluginLock: lockPath,
		Out:        outPath,
		Title:      "FBA API",
		Version:    "0.1.0",
	})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(out) error = %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(content, &doc); err != nil {
		t.Fatalf("Unmarshal(out) error = %v", err)
	}
	paths := doc["paths"].(map[string]any)
	if _, ok := paths["/api/v1/sys/users"]; !ok {
		t.Fatalf("admin path missing: %+v", paths)
	}
	components := doc["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)
	if _, ok := schemas["admin.User"]; !ok {
		t.Fatalf("admin.User schema missing: %+v", schemas)
	}
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	content, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal(%s) error = %v", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
