package main

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitUsesModuleArgument(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FBAGO_TEMPLATE_REPLACE", "")

	if err := run([]string{"init", "--dir", dir, "github.com/acme/backend"}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(content), "module github.com/acme/backend") {
		t.Fatalf("go.mod = %q, missing module name", string(content))
	}
	if !strings.Contains(string(content), "github.com/yuWorm/fba-go-admin v0.1.4") {
		t.Fatalf("go.mod = %q, missing default Admin dependency", string(content))
	}
}

func TestRunInitAcceptsDirFlagAfterModule(t *testing.T) {
	dir := t.TempDir()

	if err := run([]string{"init", "github.com/acme/backend", "--template", "basic", "--dir", dir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(content), "module github.com/acme/backend") {
		t.Fatalf("go.mod = %q, missing module name", string(content))
	}
}

func TestRunInitAcceptsTemplateFlag(t *testing.T) {
	dir := t.TempDir()

	if err := run([]string{"init", "github.com/acme/backend", "--template", "basic", "--dir", dir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	if !strings.Contains(string(content), "test ./...") {
		t.Fatalf("Makefile = %q, missing test target", string(content))
	}
}

func TestRunInitAcceptsCoreReplaceFlag(t *testing.T) {
	dir := t.TempDir()

	if err := run([]string{"init", "github.com/acme/backend", "--template", "basic", "--core-replace", "../fba-go", "--dir", dir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(content), "replace github.com/yuWorm/fba-go => ../fba-go") {
		t.Fatalf("go.mod = %q, missing explicit core replace", string(content))
	}
}

func TestRunInitHelpDocumentsFlags(t *testing.T) {
	var output bytes.Buffer
	previous := stdout
	stdout = &output
	t.Cleanup(func() {
		stdout = previous
	})

	if err := run([]string{"init", "--help"}); err != nil {
		t.Fatalf("run init --help: %v", err)
	}
	for _, want := range []string{"Create a new FBA Go project", "--template-replace", "--core-version", "--force"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("help = %q, missing %q", output.String(), want)
		}
	}
}

func TestRunInitAcceptsCoreVersionFlag(t *testing.T) {
	dir := t.TempDir()

	if err := run([]string{"init", "github.com/acme/backend", "--template", "basic", "--core-version", "v1.2.3", "--dir", dir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(content), "github.com/yuWorm/fba-go v1.2.3") {
		t.Fatalf("go.mod = %q, missing explicit core version", string(content))
	}
}

func TestRunInitRequiresModuleArgument(t *testing.T) {
	err := run([]string{"init", "--dir", t.TempDir()})
	if err == nil {
		t.Fatal("run init succeeded, want module argument error")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s), received 0") {
		t.Fatalf("error = %q, want exact argument error", err.Error())
	}
}

func TestRunWithoutArgsPrintsHelp(t *testing.T) {
	var output bytes.Buffer
	previous := stdout
	stdout = &output
	t.Cleanup(func() {
		stdout = previous
	})

	if err := run(nil); err != nil {
		t.Fatalf("run without arguments: %v", err)
	}
	for _, want := range []string{"fbago creates and maintains FBA Go projects", "Available Commands:", "template", "--help"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("help = %q, missing %q", output.String(), want)
		}
	}
}

func TestRunPluginHelpListsSubcommands(t *testing.T) {
	var output bytes.Buffer
	previous := stdout
	stdout = &output
	t.Cleanup(func() {
		stdout = previous
	})

	if err := run([]string{"plugin", "--help"}); err != nil {
		t.Fatalf("run plugin --help: %v", err)
	}
	for _, want := range []string{"Available Commands:", "outdated", "scan", "sync", "update"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("help = %q, missing %q", output.String(), want)
		}
	}
}

func TestRunSecretGeneratePrintsRequestedEntropy(t *testing.T) {
	var output bytes.Buffer
	previous := stdout
	stdout = &output
	t.Cleanup(func() {
		stdout = previous
	})

	if err := run([]string{"secret", "generate", "--bytes", "64"}); err != nil {
		t.Fatalf("run secret generate: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(output.String()))
	if err != nil {
		t.Fatalf("decode generated secret: %v", err)
	}
	if len(raw) != 64 {
		t.Fatalf("generated entropy = %d bytes, want 64", len(raw))
	}
}

func TestRunTemplateListPrintsAvailableTemplates(t *testing.T) {
	var buf bytes.Buffer
	previous := stdout
	stdout = &buf
	t.Cleanup(func() {
		stdout = previous
	})

	if err := run([]string{"template", "list"}); err != nil {
		t.Fatalf("run template list: %v", err)
	}
	if !strings.Contains(buf.String(), "basic") {
		t.Fatalf("output = %q, missing basic", buf.String())
	}
	if !strings.Contains(buf.String(), "admin") {
		t.Fatalf("output = %q, missing admin", buf.String())
	}
}

func TestRunTemplateDiffPrintsNoChanges(t *testing.T) {
	dir := t.TempDir()
	if err := run([]string{"init", "github.com/acme/backend", "--dir", dir}); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var buf bytes.Buffer
	previous := stdout
	stdout = &buf
	t.Cleanup(func() {
		stdout = previous
	})

	if err := run([]string{"template", "diff", "--dir", dir}); err != nil {
		t.Fatalf("run template diff: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "no template changes" {
		t.Fatalf("output = %q, want no template changes", buf.String())
	}
}

func TestRunPluginSyncUsesProjectDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/acme/backend\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugins.yaml"), []byte("plugins: []\n"), 0o644); err != nil {
		t.Fatalf("write plugins.yaml: %v", err)
	}

	if err := run([]string{"plugin", "sync", "--dir", dir}); err != nil {
		t.Fatalf("run plugin sync: %v", err)
	}
	for _, path := range []string{"internal/generated/fba_plugins.gen.go", "plugins.lock"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(path))); err != nil {
			t.Fatalf("generated %s: %v", path, err)
		}
	}

	var output bytes.Buffer
	previous := stdout
	stdout = &output
	t.Cleanup(func() {
		stdout = previous
	})
	if err := run([]string{"plugin", "sync", "--dir", dir, "--check"}); err != nil {
		t.Fatalf("run plugin sync --check: %v", err)
	}
	if got := strings.TrimSpace(output.String()); got != "plugin state is synchronized; dependency updates were not checked" {
		t.Fatalf("output = %q", got)
	}
}

func TestRunModuleUseSelectsLocalCheckout(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	checkoutDir := filepath.Join(root, "modules", "admin")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.MkdirAll(checkoutDir, 0o755); err != nil {
		t.Fatalf("mkdir checkout: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module github.com/acme/backend\n\ngo 1.25.0\n\nrequire github.com/yuWorm/fba-go-admin v0.1.0\n"), 0o644); err != nil {
		t.Fatalf("write project go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(checkoutDir, "go.mod"), []byte("module github.com/yuWorm/fba-go-admin\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write checkout go.mod: %v", err)
	}

	if err := run([]string{"module", "use", "--dir", projectDir, "--path", checkoutDir, "github.com/yuWorm/fba-go-admin"}); err != nil {
		t.Fatalf("run module use: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		t.Fatalf("read project go.mod: %v", err)
	}
	if !strings.Contains(string(content), "replace github.com/yuWorm/fba-go-admin =>") {
		t.Fatalf("go.mod = %q, missing local replacement", content)
	}
}
