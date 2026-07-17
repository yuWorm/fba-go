package plugin_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	fbplugin "github.com/yuWorm/fba-go/cmd/fbago/internal/plugin"
)

func TestSyncGeneratesRegistrationAndModuleAwareLock(t *testing.T) {
	dir := t.TempDir()
	coreDir, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("Abs(core) error = %v", err)
	}
	mustWrite(t, filepath.Join(dir, "go.mod"), []byte("module github.com/acme/project\n\ngo 1.25.0\n\nrequire github.com/yuWorm/fba-go v0.0.0\n\nreplace github.com/yuWorm/fba-go => "+filepath.ToSlash(coreDir)+"\n"))
	mustWrite(t, filepath.Join(dir, "plugins.yaml"), []byte(`plugins:
  - id: order
    module: github.com/acme/project/internal/app/order
    mode: auto
`))
	mustWrite(t, filepath.Join(dir, "internal/app/order/module.go"), []byte(`package order

import "github.com/yuWorm/fba-go/core/plugin"

type module struct{}

func (module) Meta() plugin.Meta { return plugin.Meta{ID: "order", Version: "0.1.0"} }
func (module) Register(plugin.Context) error { return nil }
func FBAPlugin() plugin.Module { return module{} }
`))

	if err := fbplugin.Sync(fbplugin.SyncOptions{ModuleDir: dir}); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	generated, err := os.ReadFile(filepath.Join(dir, "internal/generated/fba_plugins.gen.go"))
	if err != nil {
		t.Fatalf("ReadFile(generated) error = %v", err)
	}
	if !strings.Contains(string(generated), `plugin0 "github.com/acme/project/internal/app/order"`) {
		t.Fatalf("generated registration =\n%s", generated)
	}

	content, err := os.ReadFile(filepath.Join(dir, "plugins.lock"))
	if err != nil {
		t.Fatalf("ReadFile(lock) error = %v", err)
	}
	var lock fbplugin.PluginLock
	if err := json.Unmarshal(content, &lock); err != nil {
		t.Fatalf("Unmarshal(lock) error = %v", err)
	}
	if lock.Version != 1 || len(lock.Plugins) != 1 {
		t.Fatalf("lock = %+v", lock)
	}
	got := lock.Plugins[0]
	if got.ID != "order" || got.Module != "github.com/acme/project" || !got.Main {
		t.Fatalf("locked plugin = %+v", got)
	}
	if err := fbplugin.Sync(fbplugin.SyncOptions{ModuleDir: dir, Check: true}); err != nil {
		t.Fatalf("Sync(check) error = %v", err)
	}
}

func TestSyncCheckRejectsStaleRegistration(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "go.mod"), []byte("module github.com/acme/project\n\ngo 1.25.0\n"))
	mustWrite(t, filepath.Join(dir, "plugins.yaml"), []byte("plugins: []\n"))
	if err := fbplugin.Sync(fbplugin.SyncOptions{ModuleDir: dir}); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	mustWrite(t, filepath.Join(dir, "internal/generated/fba_plugins.gen.go"), []byte("package generated\n"))

	err := fbplugin.Sync(fbplugin.SyncOptions{ModuleDir: dir, Check: true})
	if err == nil || !strings.Contains(err.Error(), "is stale") {
		t.Fatalf("Sync(check) error = %v, want stale generated file", err)
	}
}

func TestSyncRejectsDuplicatePluginIDs(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "plugins.yaml"), []byte(`plugins:
  - id: duplicate
    module: github.com/acme/one
  - id: duplicate
    module: github.com/acme/two
`))

	err := fbplugin.Sync(fbplugin.SyncOptions{ModuleDir: dir})
	if err == nil || !strings.Contains(err.Error(), `duplicate plugin "duplicate"`) {
		t.Fatalf("Sync() error = %v, want duplicate id error", err)
	}
}
