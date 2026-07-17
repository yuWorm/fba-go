package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	coreplugin "github.com/yuWorm/fba-go/core/plugin"
)

// SyncOptions describes the project-owned plugin manifest and its generated outputs.
type SyncOptions struct {
	ModuleDir string
	Manifest  string
	Out       string
	LockOut   string
	Package   string
	Check     bool
}

// PluginLock records the Go module version that supplied each generated plugin import.
type PluginLock struct {
	Version int            `json:"version"`
	Plugins []LockedPlugin `json:"plugins"`
}

// LockedPlugin separates the importable plugin package from the Go module that versions it.
type LockedPlugin struct {
	ID            string          `json:"id"`
	Package       string          `json:"package"`
	Mode          coreplugin.Mode `json:"mode"`
	Module        string          `json:"module"`
	ModuleVersion string          `json:"module_version,omitempty"`
	Main          bool            `json:"main,omitempty"`
	Replace       *ModuleReplace  `json:"replace,omitempty"`
}

// ModuleReplace captures a local checkout or replacement module used instead of the selected version.
type ModuleReplace struct {
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
	Commit  string `json:"commit,omitempty"`
}

type listedPackage struct {
	ImportPath string        `json:"ImportPath"`
	Module     *listedModule `json:"Module"`
}

type listedModule struct {
	Path    string        `json:"Path"`
	Version string        `json:"Version"`
	Main    bool          `json:"Main"`
	Dir     string        `json:"Dir"`
	Replace *listedModule `json:"Replace"`
}

// Sync renders deterministic registration code and a module-aware lock file.
// Check mode compares the expected bytes without mutating the project.
func Sync(opts SyncOptions) error {
	moduleDir := strings.TrimSpace(opts.ModuleDir)
	if moduleDir == "" {
		moduleDir = "."
	}
	manifestPath := resolveProjectPath(moduleDir, opts.Manifest, "plugins.yaml")
	outPath := resolveProjectPath(moduleDir, opts.Out, "internal/generated/fba_plugins.gen.go")
	lockPath := resolveProjectPath(moduleDir, opts.LockOut, "plugins.lock")

	manifest, err := ReadManifest(manifestPath)
	if err != nil {
		return err
	}
	if err := validateManifest(manifest); err != nil {
		return err
	}
	result, err := Scan(ScanOptions{Modes: []string{"manifest"}, Manifest: manifestPath})
	if err != nil {
		return err
	}
	registration, err := registrationContent(opts.Package, result)
	if err != nil {
		return err
	}
	if opts.Check {
		if err := checkGeneratedFile(outPath, registration); err != nil {
			return err
		}
	} else if err := writeGeneratedFile(outPath, registration); err != nil {
		return err
	}
	if err := tidyModule(moduleDir, opts.Check); err != nil {
		return err
	}

	lock, err := resolvePluginLock(moduleDir, result)
	if err != nil {
		return err
	}
	lockContent, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	lockContent = append(lockContent, '\n')
	if opts.Check {
		return checkGeneratedFile(lockPath, lockContent)
	}
	return writeGeneratedFile(lockPath, lockContent)
}

func validateManifest(manifest Manifest) error {
	seen := make(map[string]struct{}, len(manifest.Plugins))
	for index, item := range manifest.Plugins {
		item.ID = strings.TrimSpace(item.ID)
		item.Module = strings.TrimSpace(item.Module)
		if item.ID == "" {
			return fmt.Errorf("plugins[%d].id is required", index)
		}
		if item.Module == "" {
			return fmt.Errorf("plugin %s module is required", item.ID)
		}
		if _, exists := seen[item.ID]; exists {
			return fmt.Errorf("duplicate plugin %q", item.ID)
		}
		seen[item.ID] = struct{}{}
		switch item.Mode {
		case "", coreplugin.ModeAuto, coreplugin.ModeDisabled, coreplugin.ModePureDependency:
		default:
			return fmt.Errorf("plugin %s has unsupported mode %q", item.ID, item.Mode)
		}
	}
	return nil
}

func resolvePluginLock(moduleDir string, result ScanResult) (PluginLock, error) {
	lock := PluginLock{Version: 1, Plugins: make([]LockedPlugin, 0, len(result.Plugins))}
	if len(result.Plugins) == 0 {
		return lock, nil
	}

	args := []string{"list", "-mod=mod", "-json"}
	for _, item := range result.Plugins {
		args = append(args, item.Module)
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = moduleDir
	cmd.Env = goWorkOff(os.Environ())
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return PluginLock{}, fmt.Errorf("resolve plugin modules: %w: %s", err, detail)
		}
		return PluginLock{}, fmt.Errorf("resolve plugin modules: %w", err)
	}

	packages := make(map[string]listedPackage, len(result.Plugins))
	decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	for {
		var item listedPackage
		if err := decoder.Decode(&item); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return PluginLock{}, fmt.Errorf("decode go list output: %w", err)
		}
		packages[item.ImportPath] = item
	}

	for _, item := range result.Plugins {
		listed, ok := packages[item.Module]
		if !ok || listed.Module == nil {
			return PluginLock{}, fmt.Errorf("plugin %s package %s has no Go module", item.ID, item.Module)
		}
		resolved := LockedPlugin{
			ID:            item.ID,
			Package:       item.Module,
			Mode:          item.Mode,
			Module:        listed.Module.Path,
			ModuleVersion: listed.Module.Version,
			Main:          listed.Module.Main,
		}
		if replacement := listed.Module.Replace; replacement != nil {
			path := replacement.Path
			if replacement.Dir != "" {
				path = relativeModulePath(moduleDir, replacement.Dir)
			}
			resolved.Replace = &ModuleReplace{
				Path:    path,
				Version: replacement.Version,
				Commit:  checkoutCommit(replacement.Dir),
			}
		}
		lock.Plugins = append(lock.Plugins, resolved)
	}
	sort.Slice(lock.Plugins, func(i, j int) bool { return lock.Plugins[i].ID < lock.Plugins[j].ID })
	return lock, nil
}

func resolveProjectPath(moduleDir string, value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(moduleDir, value)
}

func relativeModulePath(moduleDir string, target string) string {
	base, err := filepath.Abs(moduleDir)
	if err != nil {
		return filepath.ToSlash(target)
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return filepath.ToSlash(target)
	}
	return filepath.ToSlash(rel)
}

func checkoutCommit(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func goWorkOff(environment []string) []string {
	result := make([]string, 0, len(environment)+1)
	for _, item := range environment {
		if strings.HasPrefix(item, "GOWORK=") {
			continue
		}
		result = append(result, item)
	}
	return append(result, "GOWORK=off")
}

func tidyModule(moduleDir string, check bool) error {
	args := []string{"mod", "tidy"}
	if check {
		args = append(args, "-diff")
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = moduleDir
	cmd.Env = goWorkOff(os.Environ())
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail != "" {
			return fmt.Errorf("go mod tidy: %w: %s", err, detail)
		}
		return fmt.Errorf("go mod tidy: %w", err)
	}
	if check && len(bytes.TrimSpace(output)) != 0 {
		return fmt.Errorf("go.mod or go.sum is stale; run fbago plugin sync:\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}

func checkGeneratedFile(path string, expected []byte) error {
	actual, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("generated file %s is missing; run fbago plugin sync", path)
		}
		return err
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("generated file %s is stale; run fbago plugin sync", path)
	}
	return nil
}
