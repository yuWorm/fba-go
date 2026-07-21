package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// VersionOptions selects plugin modules declared by a project manifest.
type VersionOptions struct {
	ModuleDir string
	Manifest  string
	Targets   []string
}

// UpdateOptions controls dependency updates and the sync that follows them.
type UpdateOptions struct {
	ModuleDir string
	Manifest  string
	Out       string
	LockOut   string
	Package   string
	Targets   []string
	To        string
	DryRun    bool
}

// ModuleStatus describes the selected Go module shared by one or more plugins.
type ModuleStatus struct {
	PluginIDs []string
	Packages  []string
	Module    string
	Current   string
	Available string
	Main      bool
	Replace   *ModuleReplace
}

// ModuleUpdate is an update that was planned or applied.
type ModuleUpdate struct {
	PluginIDs []string
	Module    string
	From      string
	To        string
}

type listedVersionModule struct {
	Path    string               `json:"Path"`
	Version string               `json:"Version"`
	Update  *listedVersionModule `json:"Update"`
}

type versionCommandRunner struct {
	goCommand func(moduleDir string, args ...string) ([]byte, error)
	sync      func(SyncOptions) error
}

// Outdated resolves the current module graph and the latest compatible version
// available for each selected plugin module. It never mutates the project.
func Outdated(opts VersionOptions) ([]ModuleStatus, error) {
	return outdatedWithRunner(opts, defaultVersionCommandRunner())
}

// Update upgrades selected external plugin modules and then synchronizes the
// generated registry, go.mod/go.sum, and plugins.lock.
func Update(opts UpdateOptions) ([]ModuleUpdate, error) {
	return updateWithRunner(opts, defaultVersionCommandRunner())
}

func defaultVersionCommandRunner() versionCommandRunner {
	return versionCommandRunner{
		goCommand: runGoVersionCommand,
		sync:      Sync,
	}
}

func outdatedWithRunner(opts VersionOptions, runner versionCommandRunner) ([]ModuleStatus, error) {
	moduleDir := strings.TrimSpace(opts.ModuleDir)
	if moduleDir == "" {
		moduleDir = "."
	}
	manifestPath := resolveProjectPath(moduleDir, opts.Manifest, "plugins.yaml")
	manifest, err := ReadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if err := validateManifest(manifest); err != nil {
		return nil, err
	}
	statuses, err := resolveModuleStatuses(moduleDir, manifest, runner.goCommand)
	if err != nil {
		return nil, err
	}
	return selectModuleStatuses(statuses, opts.Targets)
}

func updateWithRunner(opts UpdateOptions, runner versionCommandRunner) ([]ModuleUpdate, error) {
	moduleDir := strings.TrimSpace(opts.ModuleDir)
	if moduleDir == "" {
		moduleDir = "."
	}
	version := strings.TrimSpace(opts.To)
	if strings.ContainsAny(version, " \t\r\n@") {
		return nil, fmt.Errorf("update version %q is invalid", opts.To)
	}
	if version != "" && len(opts.Targets) == 0 {
		return nil, fmt.Errorf("--to requires exactly one plugin or module target")
	}

	statuses, err := outdatedWithRunner(VersionOptions{
		ModuleDir: moduleDir,
		Manifest:  opts.Manifest,
		Targets:   opts.Targets,
	}, runner)
	if err != nil {
		return nil, err
	}

	external := make([]ModuleStatus, 0, len(statuses))
	for _, status := range statuses {
		if status.Main {
			if len(opts.Targets) != 0 {
				return nil, fmt.Errorf("plugin %s belongs to the project main module %s and cannot be updated", strings.Join(status.PluginIDs, ","), status.Module)
			}
			continue
		}
		if status.Replace != nil {
			return nil, fmt.Errorf("module %s is replaced by %s; run fbago module reset %s before updating", status.Module, moduleReplaceTarget(status.Replace), status.Module)
		}
		external = append(external, status)
	}
	if version != "" && len(external) != 1 {
		return nil, fmt.Errorf("--to requires exactly one external plugin module; selected %d", len(external))
	}

	updates := make([]ModuleUpdate, 0, len(external))
	for _, status := range external {
		target := status.Available
		if version != "" {
			target = version
			if version == "latest" {
				target = status.Available
			}
		}
		if target == "" || target == status.Current {
			continue
		}
		updates = append(updates, ModuleUpdate{
			PluginIDs: append([]string(nil), status.PluginIDs...),
			Module:    status.Module,
			From:      status.Current,
			To:        target,
		})
	}
	if opts.DryRun {
		return updates, nil
	}

	if len(updates) != 0 {
		args := make([]string, 1, len(updates)+1)
		args[0] = "get"
		for _, update := range updates {
			args = append(args, update.Module+"@"+update.To)
		}
		if _, err := runner.goCommand(moduleDir, args...); err != nil {
			return nil, fmt.Errorf("update plugin modules: %w", err)
		}
	}
	if err := runner.sync(SyncOptions{
		ModuleDir: moduleDir,
		Manifest:  opts.Manifest,
		Out:       opts.Out,
		LockOut:   opts.LockOut,
		Package:   opts.Package,
	}); err != nil {
		return nil, fmt.Errorf("sync updated plugins: %w", err)
	}
	return updates, nil
}

func resolveModuleStatuses(moduleDir string, manifest Manifest, runGo func(string, ...string) ([]byte, error)) ([]ModuleStatus, error) {
	if len(manifest.Plugins) == 0 {
		return []ModuleStatus{}, nil
	}
	output, err := runGo(moduleDir, "list", "-mod=readonly", "-m", "-json", "all")
	if err != nil {
		return nil, fmt.Errorf("resolve plugin modules: %w", err)
	}
	modules, err := decodeListedModules(output)
	if err != nil {
		return nil, err
	}

	byModule := make(map[string]*ModuleStatus, len(manifest.Plugins))
	for _, item := range manifest.Plugins {
		packagePath := strings.TrimSpace(item.Module)
		listed := moduleForPackage(modules, packagePath)
		if listed == nil {
			return nil, fmt.Errorf("plugin %s package %s has no Go module", item.ID, packagePath)
		}
		status := byModule[listed.Path]
		if status == nil {
			status = &ModuleStatus{
				Module:  listed.Path,
				Current: listed.Version,
				Main:    listed.Main,
			}
			if replacement := listed.Replace; replacement != nil {
				path := replacement.Path
				if replacement.Dir != "" {
					path = relativeModulePath(moduleDir, replacement.Dir)
				}
				status.Replace = &ModuleReplace{
					Path:    path,
					Version: replacement.Version,
					Commit:  checkoutCommit(replacement.Dir),
				}
			}
			byModule[listed.Path] = status
		}
		status.PluginIDs = append(status.PluginIDs, strings.TrimSpace(item.ID))
		status.Packages = append(status.Packages, packagePath)
	}

	statuses := make([]ModuleStatus, 0, len(byModule))
	queryArgs := []string{"list", "-mod=readonly", "-m", "-u", "-json"}
	for _, status := range byModule {
		sort.Strings(status.PluginIDs)
		sort.Strings(status.Packages)
		statuses = append(statuses, *status)
		if !status.Main && status.Replace == nil {
			queryArgs = append(queryArgs, status.Module)
		}
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Module < statuses[j].Module })
	if len(queryArgs) == 5 {
		return statuses, nil
	}

	output, err = runGo(moduleDir, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("query plugin module updates: %w", err)
	}
	versions, err := decodeListedVersions(output)
	if err != nil {
		return nil, err
	}
	for index := range statuses {
		status := &statuses[index]
		if status.Main || status.Replace != nil {
			continue
		}
		listed, ok := versions[status.Module]
		if !ok {
			return nil, fmt.Errorf("module update query omitted %s", status.Module)
		}
		status.Available = listed.Version
		if listed.Update != nil {
			status.Available = listed.Update.Version
		}
		if status.Available == "" {
			status.Available = status.Current
		}
	}
	return statuses, nil
}

func decodeListedModules(content []byte) ([]listedModule, error) {
	modules := make([]listedModule, 0)
	decoder := json.NewDecoder(bytes.NewReader(content))
	for {
		var item listedModule
		if err := decoder.Decode(&item); err != nil {
			if errors.Is(err, io.EOF) {
				return modules, nil
			}
			return nil, fmt.Errorf("decode Go module list: %w", err)
		}
		modules = append(modules, item)
	}
}

func moduleForPackage(modules []listedModule, packagePath string) *listedModule {
	var selected *listedModule
	for index := range modules {
		module := &modules[index]
		if packagePath != module.Path && !strings.HasPrefix(packagePath, module.Path+"/") {
			continue
		}
		// Nested Go modules own packages beneath their path even when a parent
		// module is also present in the build list.
		if selected == nil || len(module.Path) > len(selected.Path) {
			selected = module
		}
	}
	return selected
}

func decodeListedVersions(content []byte) (map[string]listedVersionModule, error) {
	versions := make(map[string]listedVersionModule)
	decoder := json.NewDecoder(bytes.NewReader(content))
	for {
		var item listedVersionModule
		if err := decoder.Decode(&item); err != nil {
			if errors.Is(err, io.EOF) {
				return versions, nil
			}
			return nil, fmt.Errorf("decode plugin module versions: %w", err)
		}
		versions[item.Path] = item
	}
}

func selectModuleStatuses(statuses []ModuleStatus, targets []string) ([]ModuleStatus, error) {
	if len(targets) == 0 {
		return statuses, nil
	}
	lookup := make(map[string]int)
	for index, status := range statuses {
		lookup[status.Module] = index
		for _, id := range status.PluginIDs {
			lookup[id] = index
		}
		for _, packagePath := range status.Packages {
			lookup[packagePath] = index
		}
	}
	selected := make(map[int]struct{}, len(targets))
	for _, target := range targets {
		target = strings.TrimSpace(target)
		index, ok := lookup[target]
		if !ok {
			return nil, fmt.Errorf("plugin or module %q is not declared in the manifest", target)
		}
		selected[index] = struct{}{}
	}
	result := make([]ModuleStatus, 0, len(selected))
	for index, status := range statuses {
		if _, ok := selected[index]; ok {
			result = append(result, status)
		}
	}
	return result, nil
}

func moduleReplaceTarget(replacement *ModuleReplace) string {
	if replacement == nil {
		return ""
	}
	if replacement.Version != "" {
		return replacement.Path + "@" + replacement.Version
	}
	return replacement.Path
}

func runGoVersionCommand(moduleDir string, args ...string) ([]byte, error) {
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
			return nil, fmt.Errorf("go %s: %w: %s", strings.Join(args, " "), err, detail)
		}
		return nil, fmt.Errorf("go %s: %w", strings.Join(args, " "), err)
	}
	return stdout.Bytes(), nil
}
