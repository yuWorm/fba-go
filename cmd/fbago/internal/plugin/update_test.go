package plugin

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestOutdatedDeduplicatesSharedPluginModule(t *testing.T) {
	dir := writeVersionManifest(t)
	recorder := newVersionCommandRecorder(t, false)

	statuses, err := outdatedWithRunner(VersionOptions{
		ModuleDir: dir,
		Targets:   []string{"config"},
	}, recorder.runner())
	if err != nil {
		t.Fatalf("Outdated() error = %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("statuses = %+v, want one shared module", statuses)
	}
	status := statuses[0]
	if status.Module != "github.com/acme/admin" || status.Current != "v1.0.0" || status.Available != "v1.1.0" {
		t.Fatalf("status = %+v", status)
	}
	if !slices.Equal(status.PluginIDs, []string{"admin", "config"}) {
		t.Fatalf("plugin IDs = %v", status.PluginIDs)
	}
}

func TestUpdateUsesOneGoGetForSharedModuleThenSyncs(t *testing.T) {
	dir := writeVersionManifest(t)
	recorder := newVersionCommandRecorder(t, false)

	updates, err := updateWithRunner(UpdateOptions{
		ModuleDir: dir,
		Targets:   []string{"config"},
	}, recorder.runner())
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(updates) != 1 || updates[0].Module != "github.com/acme/admin" || updates[0].To != "v1.1.0" {
		t.Fatalf("updates = %+v", updates)
	}
	if len(recorder.commands) != 3 || !slices.Equal(recorder.commands[2], []string{"get", "github.com/acme/admin@v1.1.0"}) {
		t.Fatalf("go commands = %v", recorder.commands)
	}
	if len(recorder.syncs) != 1 || recorder.syncs[0].ModuleDir != dir {
		t.Fatalf("sync calls = %+v", recorder.syncs)
	}
}

func TestUpdateDryRunDoesNotMutateOrSync(t *testing.T) {
	dir := writeVersionManifest(t)
	recorder := newVersionCommandRecorder(t, false)

	updates, err := updateWithRunner(UpdateOptions{ModuleDir: dir, DryRun: true}, recorder.runner())
	if err != nil {
		t.Fatalf("Update(dry-run) error = %v", err)
	}
	if len(updates) != 1 || updates[0].From != "v1.0.0" || updates[0].To != "v1.1.0" {
		t.Fatalf("updates = %+v", updates)
	}
	if len(recorder.commands) != 2 {
		t.Fatalf("go commands = %v, want only read queries", recorder.commands)
	}
	if len(recorder.syncs) != 0 {
		t.Fatalf("sync calls = %+v, want none", recorder.syncs)
	}
}

func TestUpdateRejectsReplacedModule(t *testing.T) {
	dir := writeVersionManifest(t)
	recorder := newVersionCommandRecorder(t, true)

	_, err := updateWithRunner(UpdateOptions{ModuleDir: dir, DryRun: true}, recorder.runner())
	if err == nil || !strings.Contains(err.Error(), "fbago module reset github.com/acme/admin") {
		t.Fatalf("Update(replaced) error = %v", err)
	}
	if len(recorder.commands) != 1 {
		t.Fatalf("go commands = %v, want package resolution only", recorder.commands)
	}
}

func TestUpdateToRequiresTarget(t *testing.T) {
	recorder := newVersionCommandRecorder(t, false)

	_, err := updateWithRunner(UpdateOptions{ModuleDir: t.TempDir(), To: "v1.1.0"}, recorder.runner())
	if err == nil || !strings.Contains(err.Error(), "--to requires exactly one") {
		t.Fatalf("Update(--to) error = %v", err)
	}
	if len(recorder.commands) != 0 {
		t.Fatalf("go commands = %v, want none", recorder.commands)
	}
}

type versionCommandRecorder struct {
	t            *testing.T
	graphOutput  []byte
	updateOutput []byte
	commands     [][]string
	syncs        []SyncOptions
}

func newVersionCommandRecorder(t *testing.T, replaced bool) *versionCommandRecorder {
	t.Helper()
	adminModule := &listedModule{Path: "github.com/acme/admin", Version: "v1.0.0"}
	if replaced {
		adminModule.Replace = &listedModule{Path: "../admin-checkout"}
	}
	return &versionCommandRecorder{
		t:           t,
		graphOutput: encodeJSONStream(t, adminModule),
		updateOutput: encodeJSONStream(t, listedVersionModule{
			Path:    "github.com/acme/admin",
			Version: "v1.0.0",
			Update:  &listedVersionModule{Path: "github.com/acme/admin", Version: "v1.1.0"},
		}),
	}
}

func (r *versionCommandRecorder) runner() versionCommandRunner {
	return versionCommandRunner{
		goCommand: func(_ string, args ...string) ([]byte, error) {
			r.commands = append(r.commands, append([]string(nil), args...))
			switch {
			case len(args) != 0 && args[0] == "get":
				return nil, nil
			case slices.Contains(args, "-u"):
				return r.updateOutput, nil
			default:
				return r.graphOutput, nil
			}
		},
		sync: func(opts SyncOptions) error {
			r.syncs = append(r.syncs, opts)
			return nil
		},
	}
}

func writeVersionManifest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	content := []byte(`plugins:
  - id: admin
    module: github.com/acme/admin/modules/admin
  - id: config
    module: github.com/acme/admin/modules/config
`)
	if err := os.WriteFile(filepath.Join(dir, "plugins.yaml"), content, 0o644); err != nil {
		t.Fatalf("WriteFile(manifest) error = %v", err)
	}
	return dir
}

func encodeJSONStream(t *testing.T, values ...any) []byte {
	t.Helper()
	var content bytes.Buffer
	encoder := json.NewEncoder(&content)
	for _, value := range values {
		if err := encoder.Encode(value); err != nil {
			t.Fatalf("Encode() error = %v", err)
		}
	}
	return content.Bytes()
}
