package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitUsesModuleArgument(t *testing.T) {
	dir := t.TempDir()

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
}

func TestRunInitAcceptsDirFlagAfterModule(t *testing.T) {
	dir := t.TempDir()

	if err := run([]string{"init", "github.com/acme/backend", "--dir", dir}); err != nil {
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

func TestRunInitRequiresModuleArgument(t *testing.T) {
	err := run([]string{"init", "--dir", t.TempDir()})
	if err == nil {
		t.Fatal("run init succeeded, want module argument error")
	}
	if !strings.Contains(err.Error(), "usage: fbago init <module> [--template TEMPLATE]") {
		t.Fatalf("error = %q, want fbago init usage", err.Error())
	}
}

func TestRunUsageMentionsTemplateCommand(t *testing.T) {
	err := run(nil)
	if err == nil {
		t.Fatal("run succeeded, want usage error")
	}
	if !strings.Contains(err.Error(), "template") {
		t.Fatalf("error = %q, want template command in usage", err.Error())
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
	if strings.Contains(buf.String(), "admin") {
		t.Fatalf("output = %q, should not list external admin template", buf.String())
	}
}
