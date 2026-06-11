package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const projectManifestFile = ".fbago.yaml"

type TemplateDiffOptions struct {
	Dir      string
	Template string
}

type TemplateUpdateOptions struct {
	Dir      string
	Template string
	Force    bool
	DryRun   bool
}

type TemplateChange struct {
	Status string
	Path   string
}

type TemplateDiffResult struct {
	Entries []TemplateChange
}

type TemplateUpdateResult struct {
	Entries         []TemplateChange
	ManifestUpdated bool
}

type projectManifest struct {
	Version  int                     `yaml:"version"`
	Template projectTemplateManifest `yaml:"template"`
	Managed  []managedSource         `yaml:"managed"`
}

type projectTemplateManifest struct {
	Name         string `yaml:"name,omitempty"`
	Module       string `yaml:"module,omitempty"`
	SourceModule string `yaml:"source_module,omitempty"`
	Source       string `yaml:"source,omitempty"`
	Repo         string `yaml:"repo,omitempty"`
	Ref          string `yaml:"ref,omitempty"`
	Commit       string `yaml:"commit,omitempty"`
	TemplatePath string `yaml:"template_path,omitempty"`
	CoreVersion  string `yaml:"core_version,omitempty"`
}

type managedSource struct {
	Name       string `yaml:"name"`
	Kind       string `yaml:"kind"`
	Mode       string `yaml:"mode"`
	Path       string `yaml:"path"`
	SourcePath string `yaml:"source_path"`
}

type templateSyncPlan struct {
	Changes []plannedTemplateChange
}

type plannedTemplateChange struct {
	TemplateChange
	Content  []byte
	Manifest bool
	Delete   bool
}

func DiffTemplate(opts TemplateDiffOptions) (TemplateDiffResult, error) {
	plan, err := planTemplateSync(opts.Dir, opts.Template)
	if err != nil {
		return TemplateDiffResult{}, err
	}
	return TemplateDiffResult{Entries: publicTemplateChanges(plan.Changes)}, nil
}

func UpdateTemplate(opts TemplateUpdateOptions) (TemplateUpdateResult, error) {
	plan, err := planTemplateSync(opts.Dir, opts.Template)
	if err != nil {
		return TemplateUpdateResult{}, err
	}
	result := TemplateUpdateResult{Entries: publicTemplateChanges(plan.Changes)}
	for _, change := range plan.Changes {
		if change.Manifest {
			result.ManifestUpdated = true
		}
	}
	if opts.DryRun {
		return result, nil
	}
	for _, change := range plan.Changes {
		if change.Manifest {
			continue
		}
		// Without a recorded base snapshot, a modified managed file may contain
		// project-specific business changes. Removed managed entries are equally
		// sensitive because their directories may contain local business files.
		if (change.Status == "M" || change.Status == "D") && !opts.Force {
			return result, fmt.Errorf("template update would overwrite or delete managed files; rerun with --force after reviewing the change list: %s", strings.Join(unsafeChangePaths(plan.Changes), ", "))
		}
	}
	dir := projectDir(opts.Dir)
	for _, change := range plan.Changes {
		if change.Manifest {
			continue
		}
		if err := writePlannedChange(dir, change); err != nil {
			return result, err
		}
	}
	for _, change := range plan.Changes {
		if !change.Manifest {
			continue
		}
		if err := writePlannedChange(dir, change); err != nil {
			return result, err
		}
	}
	return result, nil
}

func planTemplateSync(dir string, templateOverride string) (templateSyncPlan, error) {
	dir = projectDir(dir)
	current, err := readProjectManifest(dir)
	if err != nil {
		return templateSyncPlan{}, err
	}
	templateSpec, err := resolveTemplateSpec(current, templateOverride)
	if err != nil {
		return templateSyncPlan{}, err
	}
	bundle, err := loadTemplate(templateSpec)
	if err != nil {
		return templateSyncPlan{}, err
	}
	module := strings.TrimSpace(current.Template.Module)
	if module == "" {
		module, err = readGoModuleName(dir)
		if err != nil {
			return templateSyncPlan{}, err
		}
	}
	coreVersion := strings.TrimSpace(current.Template.CoreVersion)
	if coreVersion == "" {
		coreVersion = developmentCoreVersion
	}
	data := templateData{
		Module:         module,
		TemplateModule: bundle.TemplateModule,
		TemplateName:   bundle.TemplateName,
		TemplateSource: bundle.TemplateSource,
		TemplateRepo:   bundle.TemplateRepo,
		TemplateRef:    bundle.TemplateRef,
		TemplateCommit: bundle.TemplateCommit,
		TemplatePath:   bundle.TemplatePath,
		CoreVersion:    coreVersion,
	}
	rendered, err := renderScaffoldFiles(bundle.Files, data)
	if err != nil {
		return templateSyncPlan{}, err
	}
	next, err := readRenderedManifest(rendered)
	if err != nil {
		return templateSyncPlan{}, err
	}
	next = mergeManagedPaths(current, next)
	changes, err := planManagedChanges(dir, rendered, next.Managed)
	if err != nil {
		return templateSyncPlan{}, err
	}
	removedChanges, err := planRemovedManagedChanges(dir, current.Managed, next.Managed)
	if err != nil {
		return templateSyncPlan{}, err
	}
	changes = append(changes, removedChanges...)
	if !reflect.DeepEqual(current, next) {
		manifestContent, err := yaml.Marshal(next)
		if err != nil {
			return templateSyncPlan{}, err
		}
		if change, ok, err := compareTargetFile(dir, projectManifestFile, manifestContent, true); err != nil {
			return templateSyncPlan{}, err
		} else if ok {
			changes = append(changes, change)
		}
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})
	return templateSyncPlan{Changes: changes}, nil
}

func readProjectManifest(dir string) (projectManifest, error) {
	path := filepath.Join(dir, projectManifestFile)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return projectManifest{}, fmt.Errorf("%s not found; run fbago init with a manifest-enabled template first", path)
		}
		return projectManifest{}, err
	}
	var manifest projectManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return projectManifest{}, fmt.Errorf("read %s: %w", path, err)
	}
	if manifest.Version == 0 {
		manifest.Version = 1
	}
	if manifest.Version != 1 {
		return projectManifest{}, fmt.Errorf("%s version %d is not supported", path, manifest.Version)
	}
	return manifest, nil
}

func resolveTemplateSpec(manifest projectManifest, override string) (string, error) {
	if template := strings.TrimSpace(override); template != "" {
		return template, nil
	}
	if strings.EqualFold(strings.TrimSpace(manifest.Template.Source), "embedded") {
		if strings.TrimSpace(manifest.Template.Name) == "" {
			return "", fmt.Errorf("%s template.name is required for embedded templates", projectManifestFile)
		}
		return manifest.Template.Name, nil
	}
	if strings.TrimSpace(manifest.Template.Repo) != "" {
		return manifestRemoteTemplateSpec(manifest.Template), nil
	}
	return "", fmt.Errorf("%s does not include a template repo; pass --template to select the source template", projectManifestFile)
}

func manifestRemoteTemplateSpec(template projectTemplateManifest) string {
	spec := strings.TrimSpace(template.Repo)
	templatePath := strings.TrimSpace(template.TemplatePath)
	ref := strings.TrimSpace(template.Ref)
	if templatePath == "" && ref == "" {
		return spec
	}
	spec += "//"
	if templatePath != "." {
		spec += templatePath
	}
	if ref != "" {
		spec += "@" + ref
	}
	return spec
}

func readGoModuleName(dir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("go.mod in %s does not define module", dir)
}

func readRenderedManifest(files []renderedScaffoldFile) (projectManifest, error) {
	for _, file := range files {
		if filepath.ToSlash(file.Path) != projectManifestFile {
			continue
		}
		var manifest projectManifest
		if err := yaml.Unmarshal(file.Content, &manifest); err != nil {
			return projectManifest{}, fmt.Errorf("rendered %s is invalid: %w", projectManifestFile, err)
		}
		if manifest.Version == 0 {
			manifest.Version = 1
		}
		if manifest.Version != 1 {
			return projectManifest{}, fmt.Errorf("rendered %s version %d is not supported", projectManifestFile, manifest.Version)
		}
		return manifest, nil
	}
	return projectManifest{}, fmt.Errorf("template must render %s", projectManifestFile)
}

func mergeManagedPaths(current projectManifest, next projectManifest) projectManifest {
	currentByKey := make(map[string]managedSource, len(current.Managed))
	nextByKey := make(map[string]struct{}, len(next.Managed))
	for _, item := range current.Managed {
		currentByKey[managedKey(item)] = item
	}
	for i, item := range next.Managed {
		nextByKey[managedKey(item)] = struct{}{}
		currentItem, ok := currentByKey[managedKey(item)]
		if !ok {
			continue
		}
		if isManualManagedMode(currentItem.Mode) {
			next.Managed[i] = currentItem
			continue
		}
		if strings.TrimSpace(currentItem.Path) != "" {
			next.Managed[i].Path = currentItem.Path
		}
		if strings.TrimSpace(currentItem.Mode) != "" {
			next.Managed[i].Mode = currentItem.Mode
		}
	}
	for _, item := range current.Managed {
		if _, ok := nextByKey[managedKey(item)]; ok || !isManualManagedMode(item.Mode) {
			continue
		}
		next.Managed = append(next.Managed, item)
	}
	return next
}

func managedKey(item managedSource) string {
	return strings.TrimSpace(item.Kind) + "\x00" + strings.TrimSpace(item.Name)
}

func isManualManagedMode(mode string) bool {
	// Manual modes are project-side escape hatches: keep the manifest entry for
	// traceability but stop template diff/update from touching that path.
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "manual", "ignore", "ignored", "detached":
		return true
	default:
		return false
	}
}

func planManagedChanges(dir string, files []renderedScaffoldFile, managed []managedSource) ([]plannedTemplateChange, error) {
	changes := make([]plannedTemplateChange, 0)
	seenTargets := make(map[string]string)
	for _, item := range managed {
		if isManualManagedMode(item.Mode) {
			continue
		}
		sourceBase := item.SourcePath
		if strings.TrimSpace(sourceBase) == "" {
			sourceBase = item.Path
		}
		sourceBase, err := cleanManifestRelPath(sourceBase, "source_path")
		if err != nil {
			return nil, err
		}
		targetBase, err := cleanManifestRelPath(item.Path, "path")
		if err != nil {
			return nil, err
		}
		matched := false
		for _, file := range files {
			suffix, ok := managedFileSuffix(filepath.ToSlash(file.Path), sourceBase)
			if !ok {
				continue
			}
			matched = true
			target := targetBase
			if suffix != "" {
				target = path.Join(targetBase, suffix)
			}
			if owner, exists := seenTargets[target]; exists && owner != managedKey(item) {
				return nil, fmt.Errorf("managed target %s is declared more than once", target)
			}
			seenTargets[target] = managedKey(item)
			change, ok, err := compareTargetFile(dir, target, file.Content, false)
			if err != nil {
				return nil, err
			}
			if ok {
				changes = append(changes, change)
			}
		}
		if !matched {
			return nil, fmt.Errorf("managed source %s has no rendered files", sourceBase)
		}
	}
	return changes, nil
}

func planRemovedManagedChanges(dir string, current []managedSource, next []managedSource) ([]plannedTemplateChange, error) {
	nextByKey := make(map[string]managedSource, len(next))
	nextPaths := make([]string, 0, len(next))
	for _, item := range next {
		nextByKey[managedKey(item)] = item
		targetBase, err := cleanManifestRelPath(item.Path, "path")
		if err != nil {
			return nil, err
		}
		nextPaths = append(nextPaths, targetBase)
	}

	changes := make([]plannedTemplateChange, 0)
	for _, item := range current {
		if isManualManagedMode(item.Mode) {
			continue
		}
		if _, ok := nextByKey[managedKey(item)]; ok {
			continue
		}
		targetBase, err := cleanManifestRelPath(item.Path, "path")
		if err != nil {
			return nil, err
		}
		if overlapsAnyManagedPath(targetBase, nextPaths) {
			continue
		}
		files, err := listProjectFiles(dir, targetBase)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			changes = append(changes, plannedTemplateChange{
				TemplateChange: TemplateChange{Status: "D", Path: file},
				Delete:         true,
			})
		}
	}
	return changes, nil
}

func overlapsAnyManagedPath(target string, paths []string) bool {
	for _, item := range paths {
		if target == item || strings.HasPrefix(target, item+"/") || strings.HasPrefix(item, target+"/") {
			return true
		}
	}
	return false
}

func listProjectFiles(dir string, rel string) ([]string, error) {
	root := filepath.Join(dir, filepath.FromSlash(rel))
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return []string{rel}, nil
	}
	files := make([]string, 0)
	if err := filepath.WalkDir(root, func(item string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		file, err := filepath.Rel(dir, item)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(file))
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func cleanManifestRelPath(value string, field string) (string, error) {
	original := strings.TrimSpace(value)
	if original == "" {
		return "", fmt.Errorf("managed %s is required", field)
	}
	cleaned := path.Clean(filepath.ToSlash(original))
	if cleaned == "." || path.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("managed %s %q must be a relative project path", field, original)
	}
	return cleaned, nil
}

func managedFileSuffix(file string, sourceBase string) (string, bool) {
	if file == sourceBase {
		return "", true
	}
	prefix := sourceBase + "/"
	if strings.HasPrefix(file, prefix) {
		return strings.TrimPrefix(file, prefix), true
	}
	return "", false
}

func compareTargetFile(dir string, rel string, content []byte, manifest bool) (plannedTemplateChange, bool, error) {
	current, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		if os.IsNotExist(err) {
			return plannedTemplateChange{
				TemplateChange: TemplateChange{Status: "A", Path: rel},
				Content:        content,
				Manifest:       manifest,
			}, true, nil
		}
		return plannedTemplateChange{}, false, err
	}
	if bytes.Equal(current, content) {
		return plannedTemplateChange{}, false, nil
	}
	return plannedTemplateChange{
		TemplateChange: TemplateChange{Status: "M", Path: rel},
		Content:        content,
		Manifest:       manifest,
	}, true, nil
}

func publicTemplateChanges(changes []plannedTemplateChange) []TemplateChange {
	entries := make([]TemplateChange, 0, len(changes))
	for _, change := range changes {
		entries = append(entries, change.TemplateChange)
	}
	return entries
}

func unsafeChangePaths(changes []plannedTemplateChange) []string {
	paths := make([]string, 0)
	for _, change := range changes {
		if change.Manifest || (change.Status != "M" && change.Status != "D") {
			continue
		}
		paths = append(paths, change.Path)
	}
	sort.Strings(paths)
	return paths
}

func writePlannedChange(dir string, change plannedTemplateChange) error {
	path := filepath.Join(dir, filepath.FromSlash(change.Path))
	if change.Delete {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, change.Content, 0o644)
}

func projectDir(value string) string {
	if strings.TrimSpace(value) == "" {
		return "."
	}
	return value
}
