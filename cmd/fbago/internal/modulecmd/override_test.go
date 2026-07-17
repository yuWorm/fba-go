package modulecmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	modulecmd "github.com/yuWorm/fba-go/cmd/fbago/internal/modulecmd"
)

func TestUseAndResetLocalModuleCheckout(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	checkoutDir := filepath.Join(root, "modules", "fba-go-admin")
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/project\n\ngo 1.25.0\n\nrequire github.com/yuWorm/fba-go-admin v0.5.0\n")
	writeFile(t, filepath.Join(checkoutDir, "go.mod"), "module github.com/yuWorm/fba-go-admin\n\ngo 1.25.0\n")

	if err := modulecmd.Use(modulecmd.UseOptions{
		ProjectDir: projectDir,
		Module:     "github.com/yuWorm/fba-go-admin",
		Path:       checkoutDir,
	}); err != nil {
		t.Fatalf("Use() error = %v", err)
	}

	cmd := exec.Command("go", "list", "-m", "-f", "{{with .Replace}}{{.Dir}}{{end}}", "github.com/yuWorm/fba-go-admin")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list replacement error = %v", err)
	}
	resolved, err := filepath.EvalSymlinks(strings.TrimSpace(string(output)))
	if err != nil {
		t.Fatalf("EvalSymlinks(replacement) error = %v", err)
	}
	want, err := filepath.EvalSymlinks(checkoutDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(checkout) error = %v", err)
	}
	if resolved != want {
		t.Fatalf("replacement = %q, want %q", resolved, want)
	}

	if err := modulecmd.Reset(projectDir, "github.com/yuWorm/fba-go-admin"); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	content, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if strings.Contains(string(content), "replace github.com/yuWorm/fba-go-admin") {
		t.Fatalf("go.mod still contains replacement:\n%s", content)
	}
}

func TestUseRejectsMismatchedModulePath(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	checkoutDir := filepath.Join(root, "checkout")
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/project\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(checkoutDir, "go.mod"), "module github.com/acme/other\n\ngo 1.25.0\n")

	err := modulecmd.Use(modulecmd.UseOptions{
		ProjectDir: projectDir,
		Module:     "github.com/yuWorm/fba-go-admin",
		Path:       checkoutDir,
	})
	if err == nil || !strings.Contains(err.Error(), "declares module github.com/acme/other") {
		t.Fatalf("Use() error = %v, want module mismatch", err)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
