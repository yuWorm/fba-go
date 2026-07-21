package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"text/template"

	fbplugin "github.com/yuWorm/fba-go/cmd/fbago/internal/plugin"
	fbsecret "github.com/yuWorm/fba-go/cmd/fbago/internal/secret"
	"gopkg.in/yaml.v3"
)

//go:embed templates
var templateFS embed.FS

const defaultTemplate = "admin"
const coreModulePath = "github.com/yuWorm/fba-go"
const developmentCoreVersion = "v0.0.0"
const embeddedTemplateMetadataFile = "fbago-template.yaml"

var (
	readBuildInfo          = debug.ReadBuildInfo
	queryLatestCoreVersion = goListLatestCoreVersion
)

type InitOptions struct {
	Dir             string
	Module          string
	Template        string
	TemplateReplace string
	CoreReplace     string
	CoreVersion     string
	Force           bool
}

type scaffoldFile struct {
	Path           string
	Content        string
	Renderable     bool
	PreserveModule bool
}

type renderedScaffoldFile struct {
	Path    string
	Content []byte
}

type templateBundle struct {
	Files           []scaffoldFile
	TemplateModule  string
	TemplateVersion string
	TemplateName    string
	TemplateSource  string
	TemplateRepo    string
	TemplateRef     string
	TemplateCommit  string
	TemplatePath    string
	TemplateRoot    string
}

type remoteGitTemplate struct {
	Source   string
	CloneURL string
	Subdir   string
	Ref      string
}

type templateData struct {
	Module          string
	TemplateModule  string
	TemplateName    string
	TemplateSource  string
	TemplateRepo    string
	TemplateRef     string
	TemplateCommit  string
	TemplatePath    string
	TemplateVersion string
	TemplateReplace string
	CoreReplace     string
	CoreVersion     string
	JWTSecret       string
}

type localTemplateMetadata struct {
	Module              string   `yaml:"module"`
	Exclude             []string `yaml:"exclude"`
	PreserveModulePaths []string `yaml:"preserve_module_paths"`
}

type embeddedTemplateMetadata struct {
	Module  string `yaml:"module"`
	Version string `yaml:"version"`
}

// Local template paths usually point at a real, runnable template repository.
// Skip repository metadata and local build artifacts while keeping project files copyable.
var localTemplateSkippedDirs = map[string]struct{}{
	".cache":       {},
	".codegraph":   {},
	".git":         {},
	".hg":          {},
	".svn":         {},
	"bin":          {},
	"node_modules": {},
	"tmp":          {},
}

var localTemplateSkippedFiles = map[string]struct{}{
	".DS_Store":            {},
	".fbago-template.yaml": {},
	"Thumbs.db":            {},
}

func ListTemplates() ([]string, error) {
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func Init(opts InitOptions) error {
	module := strings.TrimSpace(opts.Module)
	if module == "" {
		return fmt.Errorf("module name is required")
	}
	if strings.ContainsAny(module, " \t\r\n") {
		return fmt.Errorf("module name must not contain whitespace")
	}
	dir := opts.Dir
	if dir == "" {
		dir = "."
	}
	templateName := strings.TrimSpace(opts.Template)
	if templateName == "" {
		templateName = defaultTemplate
	}

	bundle, err := loadTemplate(templateName)
	if err != nil {
		return err
	}
	files := bundle.Files

	coreReplace := resolveCoreReplace(opts.CoreReplace)
	coreVersion, err := resolveCoreVersion(opts.CoreVersion, coreReplace)
	if err != nil {
		return err
	}
	templateVersion, templateReplace := resolveTemplateDependency(bundle, opts.TemplateReplace)
	jwtSecret, err := secretForTemplate(files)
	if err != nil {
		return err
	}
	data := templateData{
		Module:          module,
		TemplateModule:  bundle.TemplateModule,
		TemplateName:    bundle.TemplateName,
		TemplateSource:  bundle.TemplateSource,
		TemplateRepo:    bundle.TemplateRepo,
		TemplateRef:     bundle.TemplateRef,
		TemplateCommit:  bundle.TemplateCommit,
		TemplatePath:    bundle.TemplatePath,
		TemplateVersion: templateVersion,
		TemplateReplace: templateReplace,
		CoreReplace:     coreReplace,
		CoreVersion:     coreVersion,
		JWTSecret:       jwtSecret,
	}
	renderedFiles, err := renderScaffoldFiles(files, data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	projectRoot, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer projectRoot.Close()
	if !opts.Force {
		for _, file := range renderedFiles {
			name, err := cleanRootRelativePath(file.Path)
			if err != nil {
				return err
			}
			if err := rejectRootSymlinks(projectRoot, name); err != nil {
				return err
			}
			if _, err := projectRoot.Stat(name); err == nil {
				return fmt.Errorf("%s already exists", filepath.Join(dir, name))
			} else if !os.IsNotExist(err) {
				return err
			}
		}
	}
	for _, file := range renderedFiles {
		name, err := cleanRootRelativePath(file.Path)
		if err != nil {
			return err
		}
		if err := rejectRootSymlinks(projectRoot, name); err != nil {
			return err
		}
		if err := projectRoot.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			return err
		}
		if err := projectRoot.WriteFile(name, file.Content, 0o644); err != nil {
			return err
		}
	}
	if err := rejectRootSymlinks(projectRoot, "plugins.yaml"); err != nil {
		return err
	}
	if _, err := projectRoot.Stat("plugins.yaml"); err == nil {
		return fbplugin.Sync(fbplugin.SyncOptions{ModuleDir: dir})
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func secretForTemplate(files []scaffoldFile) (string, error) {
	for _, file := range files {
		if file.Renderable && strings.Contains(file.Content, ".JWTSecret") {
			return fbsecret.Generate(fbsecret.DefaultBytes)
		}
	}
	return "", nil
}

func renderScaffoldFiles(files []scaffoldFile, data templateData) ([]renderedScaffoldFile, error) {
	renderedFiles := make([]renderedScaffoldFile, 0, len(files))
	for _, file := range files {
		source := file.Content
		if data.TemplateModule != "" && !file.PreserveModule {
			source = strings.ReplaceAll(source, data.TemplateModule, data.Module)
		}
		content := []byte(source)
		if file.Renderable {
			rendered, err := render(source, data)
			if err != nil {
				return nil, fmt.Errorf("render %s: %w", file.Path, err)
			}
			content = rendered
		}
		content = formatGoSource(file.Path, content)
		renderedFiles = append(renderedFiles, renderedScaffoldFile{
			Path:    file.Path,
			Content: content,
		})
	}
	return renderedFiles, nil
}

func resolveCoreReplace(value string) string {
	if replace := strings.TrimSpace(value); replace != "" {
		return filepath.ToSlash(replace)
	}
	if replace := strings.TrimSpace(os.Getenv("FBAGO_CORE_REPLACE")); replace != "" {
		return filepath.ToSlash(replace)
	}
	// Local development builds need a replace because the core module may not be
	// published yet; released fbago binaries should let Go resolve semver modules.
	if !isDevelopmentBuild() {
		return ""
	}
	root, err := discoverCoreModuleRoot()
	if err != nil {
		return ""
	}
	return filepath.ToSlash(root)
}

func resolveTemplateReplace(value string) string {
	if replacement := strings.TrimSpace(value); replacement != "" {
		return filepath.ToSlash(replacement)
	}
	return filepath.ToSlash(strings.TrimSpace(os.Getenv("FBAGO_TEMPLATE_REPLACE")))
}

func resolveTemplateDependency(bundle templateBundle, replacementOverride string) (version string, replace string) {
	if strings.TrimSpace(bundle.TemplateModule) == "" {
		return "", ""
	}
	if replacement := resolveTemplateReplace(replacementOverride); replacement != "" {
		return developmentCoreVersion, replacement
	}
	if bundle.TemplateSource == "local" {
		return developmentCoreVersion, filepath.ToSlash(bundle.TemplateRoot)
	}
	if bundle.TemplateSource == "embedded" {
		return strings.TrimSpace(bundle.TemplateVersion), ""
	}
	// A module in a repository subdirectory is tagged as "subdir/vX.Y.Z";
	// only the semantic-version suffix belongs in the generated go.mod.
	candidate := path.Base(strings.TrimSpace(bundle.TemplateRef))
	if strings.HasPrefix(candidate, "v") && strings.Count(candidate, ".") >= 2 {
		return candidate, ""
	}
	return developmentCoreVersion, ""
}

func resolveCoreVersion(value string, coreReplace string) (string, error) {
	version := strings.TrimSpace(value)
	if version == "" {
		version = strings.TrimSpace(os.Getenv("FBAGO_CORE_VERSION"))
	}
	switch version {
	case "":
		if strings.TrimSpace(coreReplace) != "" {
			// Go still needs a syntactically valid version even though replace makes
			// the selected version irrelevant for local development templates.
			return developmentCoreVersion, nil
		}
		if buildVersion, ok := releaseBuildCoreVersion(); ok {
			return buildVersion, nil
		}
		return developmentCoreVersion, nil
	case "latest":
		return queryLatestCoreVersion()
	default:
		return version, nil
	}
}

func releaseBuildCoreVersion() (string, bool) {
	info, ok := readBuildInfo()
	if !ok {
		return "", false
	}
	version := strings.TrimSpace(info.Main.Version)
	if version == "" || version == "(devel)" {
		return "", false
	}
	return version, true
}

func goListLatestCoreVersion() (string, error) {
	output, err := exec.Command("go", "list", "-m", "-f", "{{.Version}}", coreModulePath+"@latest").CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail != "" {
			return "", fmt.Errorf("resolve latest %s: %w: %s", coreModulePath, err, detail)
		}
		return "", fmt.Errorf("resolve latest %s: %w", coreModulePath, err)
	}
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("resolve latest %s: empty version", coreModulePath)
	}
	return version, nil
}

func isDevelopmentBuild() bool {
	info, ok := readBuildInfo()
	if !ok {
		return false
	}
	version := strings.TrimSpace(info.Main.Version)
	return version == "" || version == "(devel)"
}

func discoverCoreModuleRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("discover core module root: caller unavailable")
	}
	dir := filepath.Dir(file)
	for {
		content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(content), "module "+coreModulePath) {
			return filepath.Abs(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("discover core module root: %s not found", coreModulePath)
		}
		dir = parent
	}
}

func loadTemplate(name string) (templateBundle, error) {
	remote, ok, err := parseRemoteGitTemplate(name)
	if err != nil {
		return templateBundle{}, err
	}
	if ok {
		return loadRemoteGitTemplate(remote)
	}
	if isLocalTemplatePath(name) {
		return loadLocalTemplate(name)
	}
	if strings.Contains(name, "/") || strings.Contains(name, `\`) || name == "." || name == ".." {
		return templateBundle{}, fmt.Errorf("invalid template %q", name)
	}
	if err := ensureTemplateExists(name); err != nil {
		return templateBundle{}, err
	}
	return loadEmbeddedTemplateFiles(name)
}

func parseRemoteGitTemplate(value string) (remoteGitTemplate, bool, error) {
	source := strings.TrimSpace(value)
	if source == "" {
		return remoteGitTemplate{}, false, nil
	}
	spec := strings.TrimPrefix(source, "git+")
	if schemeIndex := strings.Index(spec, "://"); schemeIndex >= 0 {
		cloneURL, subdir, ref, err := parseURLGitTemplate(spec, schemeIndex)
		if err != nil {
			return remoteGitTemplate{}, false, err
		}
		return remoteGitTemplate{
			Source:   source,
			CloneURL: cloneURL,
			Subdir:   subdir,
			Ref:      ref,
		}, true, nil
	}
	if strings.HasPrefix(spec, "git@") {
		cloneURL, subdir, ref, err := parseSSHGitTemplate(spec)
		if err != nil {
			return remoteGitTemplate{}, false, err
		}
		return remoteGitTemplate{
			Source:   source,
			CloneURL: cloneURL,
			Subdir:   subdir,
			Ref:      ref,
		}, true, nil
	}
	cloneURL, subdir, ref, ok, err := parseHostedGitTemplate(spec)
	if err != nil || !ok {
		return remoteGitTemplate{}, ok, err
	}
	return remoteGitTemplate{
		Source:   source,
		CloneURL: cloneURL,
		Subdir:   subdir,
		Ref:      ref,
	}, true, nil
}

func parseURLGitTemplate(spec string, schemeIndex int) (string, string, string, error) {
	searchStart := schemeIndex + len("://")
	separator := strings.Index(spec[searchStart:], "//")
	if separator < 0 {
		return spec, ".", "", nil
	}
	separator += searchStart
	cloneURL := spec[:separator]
	subdir, ref, err := splitRemoteSubdirRef(spec[separator+2:])
	if err != nil {
		return "", "", "", err
	}
	return cloneURL, subdir, ref, nil
}

func parseSSHGitTemplate(spec string) (string, string, string, error) {
	separator := strings.Index(spec, "//")
	if separator < 0 {
		return spec, ".", "", nil
	}
	subdir, ref, err := splitRemoteSubdirRef(spec[separator+2:])
	if err != nil {
		return "", "", "", err
	}
	return spec[:separator], subdir, ref, nil
}

func parseHostedGitTemplate(spec string) (string, string, string, bool, error) {
	base, ref := splitRef(spec)
	parts := strings.Split(base, "/")
	if len(parts) < 3 || !strings.Contains(parts[0], ".") {
		return "", "", "", false, nil
	}
	repo := strings.Join(parts[:3], "/")
	cloneURL := "https://" + repo
	if !strings.HasSuffix(cloneURL, ".git") {
		cloneURL += ".git"
	}
	subdir := "."
	if len(parts) > 3 {
		subdir = strings.Join(parts[3:], "/")
	}
	subdir, err := cleanRemoteSubdir(subdir)
	if err != nil {
		return "", "", "", false, err
	}
	return cloneURL, subdir, ref, true, nil
}

func splitRemoteSubdirRef(value string) (string, string, error) {
	subdir, ref := splitRef(value)
	subdir, err := cleanRemoteSubdir(subdir)
	if err != nil {
		return "", "", err
	}
	return subdir, ref, nil
}

func splitRef(value string) (string, string) {
	index := strings.LastIndex(value, "@")
	if index < 0 {
		return value, ""
	}
	return value[:index], value[index+1:]
}

func cleanRemoteSubdir(subdir string) (string, error) {
	subdir = strings.TrimSpace(subdir)
	if subdir == "" {
		return ".", nil
	}
	cleaned := path.Clean(filepath.ToSlash(subdir))
	if cleaned == "." {
		return ".", nil
	}
	if path.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("invalid remote template subdir %q", subdir)
	}
	return cleaned, nil
}

func loadRemoteGitTemplate(spec remoteGitTemplate) (templateBundle, error) {
	checkout, cleanup, err := cloneRemoteGitTemplate(spec)
	if err != nil {
		return templateBundle{}, err
	}
	defer cleanup()

	root := checkout
	if spec.Subdir != "" && spec.Subdir != "." {
		root = filepath.Join(root, filepath.FromSlash(spec.Subdir))
	}
	bundle, err := loadLocalTemplate(root)
	if err != nil {
		return templateBundle{}, err
	}
	bundle.TemplateSource = "remote"
	bundle.TemplateRepo = spec.CloneURL
	bundle.TemplateRef = spec.Ref
	if spec.Subdir != "" && spec.Subdir != "." {
		bundle.TemplatePath = filepath.ToSlash(spec.Subdir)
	}
	return bundle, nil
}

func cloneRemoteGitTemplate(spec remoteGitTemplate) (string, func(), error) {
	cacheRoot, err := templateCacheRoot()
	if err != nil {
		return "", nil, err
	}
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return "", nil, err
	}
	checkout, err := os.MkdirTemp(cacheRoot, "checkout-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_ = os.RemoveAll(checkout)
	}

	args := []string{"clone", "--quiet", "--depth", "1"}
	if spec.Ref != "" {
		args = append(args, "--branch", spec.Ref)
	}
	args = append(args, spec.CloneURL, checkout)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("clone template %q: %w: %s", spec.Source, err, strings.TrimSpace(string(output)))
	}
	return checkout, cleanup, nil
}

func templateCacheRoot() (string, error) {
	if root := strings.TrimSpace(os.Getenv("FBAGO_TEMPLATE_CACHE_DIR")); root != "" {
		return root, nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "fbago", "templates"), nil
	}
	return filepath.Join(cache, "fbago", "templates"), nil
}

func isLocalTemplatePath(value string) bool {
	if filepath.IsAbs(value) {
		return true
	}
	return strings.HasPrefix(value, ".") ||
		strings.Contains(value, "/") ||
		strings.Contains(value, `\`)
}

func ensureTemplateExists(name string) error {
	root := path.Join("templates", name)
	info, err := fs.Stat(templateFS, root)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("unknown template %q", name)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("unknown template %q", name)
	}
	return nil
}

func loadEmbeddedTemplateFiles(name string) (templateBundle, error) {
	root := path.Join("templates", name)
	metadata, err := readEmbeddedTemplateMetadata(root)
	if err != nil {
		return templateBundle{}, err
	}
	files := make([]scaffoldFile, 0)
	if err := fs.WalkDir(templateFS, root, func(item string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := templateFS.ReadFile(item)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, item)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == embeddedTemplateMetadataFile {
			return nil
		}
		// Files ending in .tmpl participate in module-name rendering and drop the suffix.
		// Root env templates map to dotfiles because go:embed ignores hidden files.
		target, renderable := targetPath(rel)
		files = append(files, scaffoldFile{
			Path:       target,
			Content:    string(content),
			Renderable: renderable,
		})
		return nil
	}); err != nil {
		return templateBundle{}, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	return templateBundle{
		Files:           files,
		TemplateModule:  metadata.Module,
		TemplateVersion: metadata.Version,
		TemplateName:    name,
		TemplateSource:  "embedded",
		TemplatePath:    name,
	}, nil
}

func readEmbeddedTemplateMetadata(root string) (embeddedTemplateMetadata, error) {
	metadataPath := path.Join(root, embeddedTemplateMetadataFile)
	content, err := templateFS.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return embeddedTemplateMetadata{}, nil
		}
		return embeddedTemplateMetadata{}, err
	}
	var metadata embeddedTemplateMetadata
	if err := yaml.Unmarshal(content, &metadata); err != nil {
		return embeddedTemplateMetadata{}, fmt.Errorf("read embedded template metadata %s: %w", metadataPath, err)
	}
	metadata.Module = strings.TrimSpace(metadata.Module)
	metadata.Version = strings.TrimSpace(metadata.Version)
	if metadata.Module == "" {
		return embeddedTemplateMetadata{}, fmt.Errorf("embedded template metadata %s must define module", metadataPath)
	}
	if strings.ContainsAny(metadata.Module, " \t\r\n") {
		return embeddedTemplateMetadata{}, fmt.Errorf("embedded template metadata %s module must not contain whitespace", metadataPath)
	}
	if !strings.HasPrefix(metadata.Version, "v") || strings.Count(metadata.Version, ".") < 2 {
		return embeddedTemplateMetadata{}, fmt.Errorf("embedded template metadata %s must define a semantic version", metadataPath)
	}
	return metadata, nil
}

func loadLocalTemplate(root string) (templateBundle, error) {
	info, err := os.Lstat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return templateBundle{}, fmt.Errorf("template path %q does not exist", root)
		}
		return templateBundle{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return templateBundle{}, fmt.Errorf("template path %q must not be a symbolic link", root)
	}
	if !info.IsDir() {
		return templateBundle{}, fmt.Errorf("template path %q is not a directory", root)
	}
	templateRoot, err := os.OpenRoot(root)
	if err != nil {
		return templateBundle{}, err
	}
	defer templateRoot.Close()

	metadata, err := readLocalTemplateMetadata(templateRoot, root)
	if err != nil {
		return templateBundle{}, err
	}

	files := make([]scaffoldFile, 0)
	if err := fs.WalkDir(templateRoot.FS(), ".", func(item string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(item)
		if item != "." && templatePathMatches(rel, metadata.Exclude) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("template entry %q must not be a symbolic link", filepath.Join(root, item))
		}
		if entry.IsDir() {
			if item != "." && shouldSkipLocalTemplateDir(entry.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if shouldSkipLocalTemplateFile(entry.Name()) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("template entry %q must be a regular file", filepath.Join(root, item))
		}
		content, err := templateRoot.ReadFile(item)
		if err != nil {
			return err
		}
		target, renderable := targetPath(rel)
		files = append(files, scaffoldFile{
			Path:           target,
			Content:        string(content),
			Renderable:     renderable,
			PreserveModule: templatePathMatches(rel, metadata.PreserveModulePaths),
		})
		return nil
	}); err != nil {
		return templateBundle{}, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	repo, commit, templatePath := localTemplateGitMetadata(root)
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return templateBundle{}, err
	}
	return templateBundle{
		Files:          files,
		TemplateModule: metadata.Module,
		TemplateName:   filepath.Base(root),
		TemplateSource: "local",
		TemplateRepo:   repo,
		TemplateCommit: commit,
		TemplatePath:   templatePath,
		TemplateRoot:   resolvedRoot,
	}, nil
}

func localTemplateGitMetadata(root string) (repo string, commit string, templatePath string) {
	// Template origin metadata is best-effort: ordinary local directories used in
	// tests or private templates may not be Git checkouts, but official templates
	// should record enough information for future diff/update commands.
	if absRoot, err := filepath.Abs(root); err == nil {
		root = absRoot
	}
	repo = gitOutput(root, "config", "--get", "remote.origin.url")
	commit = gitOutput(root, "rev-parse", "HEAD")
	gitRoot := gitOutput(root, "rev-parse", "--show-toplevel")
	if gitRoot == "" {
		return repo, commit, ""
	}
	resolvedGitRoot, err := filepath.EvalSymlinks(gitRoot)
	if err == nil {
		gitRoot = resolvedGitRoot
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err == nil {
		root = resolvedRoot
	}
	rel, err := filepath.Rel(gitRoot, root)
	if err != nil || rel == "." {
		return repo, commit, "."
	}
	return repo, commit, filepath.ToSlash(rel)
}

func gitOutput(root string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func readLocalTemplateMetadata(root *os.Root, displayRoot string) (localTemplateMetadata, error) {
	metadataPath := filepath.Join(displayRoot, ".fbago-template.yaml")
	content, err := root.ReadFile(".fbago-template.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return localTemplateMetadata{}, nil
		}
		return localTemplateMetadata{}, err
	}
	var metadata localTemplateMetadata
	if err := yaml.Unmarshal(content, &metadata); err != nil {
		return localTemplateMetadata{}, fmt.Errorf("read template metadata %s: %w", metadataPath, err)
	}
	metadata.Module = strings.TrimSpace(metadata.Module)
	if metadata.Module == "" {
		return localTemplateMetadata{}, fmt.Errorf("template metadata %s must define module", metadataPath)
	}
	if strings.ContainsAny(metadata.Module, " \t\r\n") {
		return localTemplateMetadata{}, fmt.Errorf("template metadata %s module must not contain whitespace", metadataPath)
	}
	if err := validateTemplateMetadataPaths(metadataPath, metadata.Exclude); err != nil {
		return localTemplateMetadata{}, err
	}
	if err := validateTemplateMetadataPaths(metadataPath, metadata.PreserveModulePaths); err != nil {
		return localTemplateMetadata{}, err
	}
	return metadata, nil
}

func validateTemplateMetadataPaths(metadataPath string, values []string) error {
	for _, value := range values {
		cleaned := path.Clean(strings.TrimSpace(filepath.ToSlash(value)))
		if cleaned == "." || cleaned == "" || strings.HasPrefix(cleaned, "../") || path.IsAbs(cleaned) {
			return fmt.Errorf("template metadata %s contains invalid path %q", metadataPath, value)
		}
	}
	return nil
}

func templatePathMatches(rel string, values []string) bool {
	rel = path.Clean(filepath.ToSlash(rel))
	for _, value := range values {
		cleaned := path.Clean(strings.TrimSpace(filepath.ToSlash(value)))
		if rel == cleaned || strings.HasPrefix(rel, cleaned+"/") {
			return true
		}
	}
	return false
}

func formatGoSource(target string, content []byte) []byte {
	if filepath.Ext(target) != ".go" {
		return content
	}
	formatted, err := format.Source(content)
	if err != nil {
		// Templates may carry intentionally incomplete Go snippets in rare cases.
		// Keep scaffold generation tolerant and let the generated project tests fail if needed.
		return content
	}
	return formatted
}

func shouldSkipLocalTemplateDir(name string) bool {
	_, ok := localTemplateSkippedDirs[name]
	return ok
}

func shouldSkipLocalTemplateFile(name string) bool {
	_, ok := localTemplateSkippedFiles[name]
	return ok
}

func targetPath(rel string) (string, bool) {
	if rel == "env.tmpl" {
		return ".env", true
	}
	if rel == "env.example.tmpl" {
		return ".env.example", true
	}
	if rel == "fbago.yaml.tmpl" {
		return ".fbago.yaml", true
	}
	if rel == "gitignore.tmpl" {
		return ".gitignore", true
	}
	if strings.HasSuffix(rel, ".tmpl") {
		return strings.TrimSuffix(rel, ".tmpl"), true
	}
	return rel, false
}

func render(source string, data templateData) ([]byte, error) {
	// Use delimiters that do not collide with ordinary Go composite literals
	// such as []plugin.Dependency{{ID: "admin"}} inside copied source templates.
	tmpl, err := template.New("scaffold").Delims("[[", "]]").Parse(source)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
