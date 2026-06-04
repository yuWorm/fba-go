package main

import (
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

func TestRunInitRequiresModuleArgument(t *testing.T) {
	err := run([]string{"init", "--dir", t.TempDir()})
	if err == nil {
		t.Fatal("run init succeeded, want module argument error")
	}
	if !strings.Contains(err.Error(), "usage: fbago init <module>") {
		t.Fatalf("error = %q, want fbago init usage", err.Error())
	}
}
