package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuWorm/fba-go/cmd/fbago/internal/scaffold"
)

func TestInitWritesBackendScaffoldWithModuleName(t *testing.T) {
	dir := t.TempDir()

	if err := scaffold.Init(scaffold.InitOptions{
		Dir:    dir,
		Module: "github.com/acme/backend",
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	assertFileContains(t, filepath.Join(dir, "go.mod"), "module github.com/acme/backend")
	assertFileContains(t, filepath.Join(dir, "cmd/api/main.go"), `"github.com/acme/backend/internal/app"`)
	assertFileContains(t, filepath.Join(dir, "internal/app/register.go"), `"github.com/acme/backend/internal/app/health"`)
	assertFileContains(t, filepath.Join(dir, "internal/app/health/module.go"), `ID:          "health"`)
	assertFileContains(t, filepath.Join(dir, ".env"), "FASTAPI_API_V1_PATH=/api/v1")
	assertFileContains(t, filepath.Join(dir, "README.md"), "fbago init github.com/acme/backend")
}

func TestInitRejectsMissingModule(t *testing.T) {
	err := scaffold.Init(scaffold.InitOptions{Dir: t.TempDir()})
	if err == nil {
		t.Fatal("Init() succeeded, want missing module error")
	}
	if !strings.Contains(err.Error(), "module name is required") {
		t.Fatalf("error = %q, want module name error", err.Error())
	}
}

func TestInitDoesNotOverwriteExistingGoMod(t *testing.T) {
	dir := t.TempDir()
	goMod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module existing\n"), 0o644); err != nil {
		t.Fatalf("write existing go.mod: %v", err)
	}

	err := scaffold.Init(scaffold.InitOptions{
		Dir:    dir,
		Module: "github.com/acme/backend",
	})
	if err == nil {
		t.Fatal("Init() succeeded, want overwrite error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %q, want already exists", err.Error())
	}
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("%s = %q, missing %q", path, string(content), want)
	}
}
