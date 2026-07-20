package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuWorm/fba-go/cmd/fbago/internal/scaffold"
)

func TestDiffTemplateReportsManagedAddAndModify(t *testing.T) {
	dir := t.TempDir()
	templateDir := t.TempDir()
	writeSyncTemplate(t, templateDir, "admin v1\n", false)

	if err := scaffold.Init(scaffold.InitOptions{
		Dir:      dir,
		Module:   "github.com/acme/backend",
		Template: templateDir,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	writeSyncTemplate(t, templateDir, "admin v2\n", true)
	result, err := scaffold.DiffTemplate(scaffold.TemplateDiffOptions{
		Dir:      dir,
		Template: templateDir,
	})
	if err != nil {
		t.Fatalf("DiffTemplate() error = %v", err)
	}

	assertTemplateChanges(t, result.Entries, []string{
		"M .fbago.yaml",
		"M internal/app/admin/version.txt",
		"A internal/app/notice/new.txt",
	})
}

func TestUpdateTemplateRefusesModifiedManagedFilesWithoutForce(t *testing.T) {
	dir := t.TempDir()
	templateDir := t.TempDir()
	writeSyncTemplate(t, templateDir, "admin v1\n", false)

	if err := scaffold.Init(scaffold.InitOptions{
		Dir:      dir,
		Module:   "github.com/acme/backend",
		Template: templateDir,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	writeSyncTemplate(t, templateDir, "admin v2\n", false)
	result, err := scaffold.UpdateTemplate(scaffold.TemplateUpdateOptions{
		Dir:      dir,
		Template: templateDir,
	})
	if err == nil {
		t.Fatal("UpdateTemplate() succeeded, want overwrite guard")
	}
	if !strings.Contains(err.Error(), "would overwrite or delete managed files") {
		t.Fatalf("error = %q, want overwrite guard", err.Error())
	}
	assertTemplateChanges(t, result.Entries, []string{
		"M internal/app/admin/version.txt",
	})
	assertFileContains(t, filepath.Join(dir, "internal/app/admin/version.txt"), "admin v1")
}

func TestUpdateTemplateAddsNewManagedFilesWithoutForce(t *testing.T) {
	dir := t.TempDir()
	templateDir := t.TempDir()
	writeSyncTemplate(t, templateDir, "admin v1\n", false)

	if err := scaffold.Init(scaffold.InitOptions{
		Dir:      dir,
		Module:   "github.com/acme/backend",
		Template: templateDir,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	writeSyncTemplate(t, templateDir, "admin v1\n", true)
	result, err := scaffold.UpdateTemplate(scaffold.TemplateUpdateOptions{
		Dir:      dir,
		Template: templateDir,
	})
	if err != nil {
		t.Fatalf("UpdateTemplate() error = %v", err)
	}
	assertTemplateChanges(t, result.Entries, []string{
		"M .fbago.yaml",
		"A internal/app/notice/new.txt",
	})
	assertFileContains(t, filepath.Join(dir, "internal/app/notice/new.txt"), "notice v1")
	assertFileContains(t, filepath.Join(dir, ".fbago.yaml"), "name: notice")
}

func TestUpdateTemplateForceOverwritesManagedFiles(t *testing.T) {
	dir := t.TempDir()
	templateDir := t.TempDir()
	writeSyncTemplate(t, templateDir, "admin v1\n", false)

	if err := scaffold.Init(scaffold.InitOptions{
		Dir:      dir,
		Module:   "github.com/acme/backend",
		Template: templateDir,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	writeSyncTemplate(t, templateDir, "admin v2\n", false)
	if _, err := scaffold.UpdateTemplate(scaffold.TemplateUpdateOptions{
		Dir:      dir,
		Template: templateDir,
		Force:    true,
	}); err != nil {
		t.Fatalf("UpdateTemplate() error = %v", err)
	}
	assertFileContains(t, filepath.Join(dir, "internal/app/admin/version.txt"), "admin v2")
}

func TestUpdateTemplateCannotWriteThroughDestinationSymbolicLink(t *testing.T) {
	dir := t.TempDir()
	templateDir := t.TempDir()
	writeSyncTemplate(t, templateDir, "admin v1\n", false)
	if err := scaffold.Init(scaffold.InitOptions{
		Dir:      dir,
		Module:   "github.com/acme/backend",
		Template: templateDir,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if err := os.RemoveAll(filepath.Join(dir, "internal")); err != nil {
		t.Fatalf("remove managed directory: %v", err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(dir, "internal")); err != nil {
		t.Skipf("symbolic links are unavailable: %v", err)
	}
	writeSyncTemplate(t, templateDir, "admin v2\n", false)
	_, err := scaffold.UpdateTemplate(scaffold.TemplateUpdateOptions{
		Dir:      dir,
		Template: templateDir,
		Force:    true,
	})
	if err == nil {
		t.Fatal("UpdateTemplate() succeeded, want destination symbolic-link rejection")
	}
	if !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("error = %q, want symbolic-link rejection", err.Error())
	}
	assertFileNotExists(t, filepath.Join(outside, "app/admin/version.txt"))
}

func TestDiffTemplateSkipsManualManagedEntries(t *testing.T) {
	dir := t.TempDir()
	templateDir := t.TempDir()
	writeSyncTemplate(t, templateDir, "admin v1\n", false)

	if err := scaffold.Init(scaffold.InitOptions{
		Dir:      dir,
		Module:   "github.com/acme/backend",
		Template: templateDir,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	manifestPath := filepath.Join(dir, ".fbago.yaml")
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	manifest = []byte(strings.Replace(string(manifest), "mode: source", "mode: manual", 1))
	if err := os.WriteFile(manifestPath, manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	writeSyncTemplate(t, templateDir, "admin v2\n", false)
	result, err := scaffold.DiffTemplate(scaffold.TemplateDiffOptions{
		Dir:      dir,
		Template: templateDir,
	})
	if err != nil {
		t.Fatalf("DiffTemplate() error = %v", err)
	}
	assertTemplateChanges(t, result.Entries, nil)
}

func TestUpdateTemplateRequiresForceForRemovedManagedFiles(t *testing.T) {
	dir := t.TempDir()
	templateDir := t.TempDir()
	nextTemplateDir := t.TempDir()
	writeSyncTemplate(t, templateDir, "admin v1\n", false)
	writeRemovedSyncTemplate(t, nextTemplateDir)

	if err := scaffold.Init(scaffold.InitOptions{
		Dir:      dir,
		Module:   "github.com/acme/backend",
		Template: templateDir,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	result, err := scaffold.UpdateTemplate(scaffold.TemplateUpdateOptions{
		Dir:      dir,
		Template: nextTemplateDir,
	})
	if err == nil {
		t.Fatal("UpdateTemplate() succeeded, want delete guard")
	}
	assertTemplateChanges(t, result.Entries, []string{
		"M .fbago.yaml",
		"D internal/app/admin/version.txt",
	})
	assertFileContains(t, filepath.Join(dir, "internal/app/admin/version.txt"), "admin v1")

	if _, err := scaffold.UpdateTemplate(scaffold.TemplateUpdateOptions{
		Dir:      dir,
		Template: nextTemplateDir,
		Force:    true,
	}); err != nil {
		t.Fatalf("UpdateTemplate(force) error = %v", err)
	}
	assertFileNotExists(t, filepath.Join(dir, "internal/app/admin/version.txt"))
	assertFileNotContains(t, filepath.Join(dir, ".fbago.yaml"), "name: admin")

	diff, err := scaffold.DiffTemplate(scaffold.TemplateDiffOptions{
		Dir:      dir,
		Template: nextTemplateDir,
	})
	if err != nil {
		t.Fatalf("DiffTemplate() after removal error = %v", err)
	}
	assertTemplateChanges(t, diff.Entries, nil)
}

func writeSyncTemplate(t *testing.T, root string, adminContent string, includeNotice bool) {
	t.Helper()
	writeTemplateFile(t, root, ".fbago-template.yaml", "module: github.com/acme/template\n")
	writeTemplateFile(t, root, "go.mod", "module github.com/acme/template\n\ngo 1.25.0\n")
	manifest := `version: 1

template:
  name: sync
  module: [[ .Module ]]
  source_module: [[ .TemplateModule ]]
  source: [[ .TemplateSource ]]
  template_path: [[ .TemplatePath ]]
  core_version: [[ .CoreVersion ]]

managed:
  - name: admin
    kind: app
    mode: source
    path: internal/app/admin
    source_path: internal/app/admin
`
	if includeNotice {
		manifest += `
  - name: notice
    kind: app
    mode: source
    path: internal/app/notice
    source_path: internal/app/notice
`
		writeTemplateFile(t, root, "internal/app/notice/new.txt", "notice v1\n")
	}
	writeTemplateFile(t, root, "fbago.yaml.tmpl", manifest)
	writeTemplateFile(t, root, "internal/app/admin/version.txt.tmpl", adminContent)
}

func writeRemovedSyncTemplate(t *testing.T, root string) {
	t.Helper()
	writeTemplateFile(t, root, ".fbago-template.yaml", "module: github.com/acme/template\n")
	writeTemplateFile(t, root, "go.mod", "module github.com/acme/template\n\ngo 1.25.0\n")
	writeTemplateFile(t, root, "fbago.yaml.tmpl", `version: 1

template:
  name: sync
  module: [[ .Module ]]
  source_module: [[ .TemplateModule ]]
  source: [[ .TemplateSource ]]
  template_path: [[ .TemplatePath ]]
  core_version: [[ .CoreVersion ]]

managed: []
`)
}

func assertTemplateChanges(t *testing.T, got []scaffold.TemplateChange, want []string) {
	t.Helper()
	lines := make([]string, 0, len(got))
	for _, entry := range got {
		lines = append(lines, entry.Status+" "+entry.Path)
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("changes = %q, want %q", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}
