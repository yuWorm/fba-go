package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	fbcontract "github.com/yuWorm/fba-go/cmd/fbago/internal/contract"
	fbmodule "github.com/yuWorm/fba-go/cmd/fbago/internal/modulecmd"
	fbplugin "github.com/yuWorm/fba-go/cmd/fbago/internal/plugin"
	"github.com/yuWorm/fba-go/cmd/fbago/internal/scaffold"
	fbswagger "github.com/yuWorm/fba-go/cmd/fbago/internal/swagger"
)

var stdout io.Writer = os.Stdout

const initUsage = "usage: fbago init <module> [--template TEMPLATE] [--template-replace PATH] [--dir DIR] [--core-replace PATH] [--core-version VERSION|latest]"
const templateDiffUsage = "usage: fbago template diff [--dir DIR] [--template TEMPLATE]"
const templateUpdateUsage = "usage: fbago template update [--dir DIR] [--template TEMPLATE] [--dry-run] [--force]"
const pluginSyncUsage = "usage: fbago plugin sync [--dir DIR] [--manifest FILE] [--out FILE] [--lock-out FILE] [--check]"
const moduleUseUsage = "usage: fbago module use [--dir DIR] --path PATH <module>"
const moduleResetUsage = "usage: fbago module reset [--dir DIR] <module>"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}
	if args[0] == "init" {
		return runInit(args[1:])
	}
	if len(args) < 2 {
		return usage()
	}
	switch args[0] + " " + args[1] {
	case "plugin scan":
		return runPluginScan(args[2:])
	case "plugin sync":
		return runPluginSync(args[2:])
	case "plugin outdated":
		return runPluginOutdated(args[2:])
	case "plugin update":
		return runPluginUpdate(args[2:])
	case "module use":
		return runModuleUse(args[2:])
	case "module reset":
		return runModuleReset(args[2:])
	case "swagger scan":
		return runSwaggerScan(args[2:])
	case "contract snapshot":
		return runContractSnapshot(args[2:])
	case "contract test":
		return runContractTest(args[2:])
	case "template list":
		return runTemplateList(args[2:])
	case "template diff":
		return runTemplateDiff(args[2:])
	case "template update":
		return runTemplateUpdate(args[2:])
	default:
		return fmt.Errorf("unknown command %s %s", args[0], args[1])
	}
}

func usage() error {
	return fmt.Errorf("usage: fbago <init|template|plugin|module|swagger|contract> [command]")
}

func runInit(args []string) error {
	opts, err := parseInitArgs(args)
	if err != nil {
		return err
	}
	if opts.Module == "" {
		return fmt.Errorf(initUsage)
	}
	return scaffold.Init(opts)
}

func runTemplateList(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: fbago template list")
	}
	templates, err := scaffold.ListTemplates()
	if err != nil {
		return err
	}
	for _, template := range templates {
		fmt.Fprintln(stdout, template)
	}
	return nil
}

func runTemplateDiff(args []string) error {
	fs := flag.NewFlagSet("template diff", flag.ContinueOnError)
	dir := fs.String("dir", ".", "project directory")
	template := fs.String("template", "", "template source override")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf(templateDiffUsage)
	}
	result, err := scaffold.DiffTemplate(scaffold.TemplateDiffOptions{
		Dir:      *dir,
		Template: *template,
	})
	if err != nil {
		return err
	}
	printTemplateChanges(result.Entries)
	return nil
}

func runTemplateUpdate(args []string) error {
	fs := flag.NewFlagSet("template update", flag.ContinueOnError)
	dir := fs.String("dir", ".", "project directory")
	template := fs.String("template", "", "template source override")
	dryRun := fs.Bool("dry-run", false, "show changes without writing")
	force := fs.Bool("force", false, "overwrite modified managed files")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf(templateUpdateUsage)
	}
	result, err := scaffold.UpdateTemplate(scaffold.TemplateUpdateOptions{
		Dir:      *dir,
		Template: *template,
		DryRun:   *dryRun,
		Force:    *force,
	})
	if err != nil && len(result.Entries) == 0 {
		return err
	}
	printTemplateChanges(result.Entries)
	return err
}

func printTemplateChanges(entries []scaffold.TemplateChange) {
	if len(entries) == 0 {
		fmt.Fprintln(stdout, "no template changes")
		return
	}
	for _, entry := range entries {
		fmt.Fprintf(stdout, "%s %s\n", entry.Status, entry.Path)
	}
}

func parseInitArgs(args []string) (scaffold.InitOptions, error) {
	opts := scaffold.InitOptions{Dir: "."}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dir", "-dir":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("missing value for %s", arg)
			}
			opts.Dir = args[i]
		case "--force", "-force":
			opts.Force = true
		case "--template", "-template":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("missing value for %s", arg)
			}
			opts.Template = args[i]
		case "--template-replace", "-template-replace":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("missing value for %s", arg)
			}
			opts.TemplateReplace = args[i]
		case "--core-replace", "-core-replace":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("missing value for %s", arg)
			}
			opts.CoreReplace = args[i]
		case "--core-version", "-core-version":
			i++
			if i >= len(args) {
				return opts, fmt.Errorf("missing value for %s", arg)
			}
			opts.CoreVersion = args[i]
		default:
			if strings.HasPrefix(arg, "-") {
				return opts, fmt.Errorf("unknown init flag %s", arg)
			}
			if opts.Module != "" {
				return opts, fmt.Errorf(initUsage)
			}
			opts.Module = arg
		}
	}
	return opts, nil
}

func runModuleUse(args []string) error {
	fs := flag.NewFlagSet("module use", flag.ContinueOnError)
	projectDir := fs.String("dir", ".", "project module directory")
	localPath := fs.String("path", "", "local module checkout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 || strings.TrimSpace(*localPath) == "" {
		return fmt.Errorf(moduleUseUsage)
	}
	return fbmodule.Use(fbmodule.UseOptions{
		ProjectDir: *projectDir,
		Module:     fs.Arg(0),
		Path:       *localPath,
	})
}

func runModuleReset(args []string) error {
	fs := flag.NewFlagSet("module reset", flag.ContinueOnError)
	projectDir := fs.String("dir", ".", "project module directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf(moduleResetUsage)
	}
	return fbmodule.Reset(*projectDir, fs.Arg(0))
}

func runPluginSync(args []string) error {
	fs := flag.NewFlagSet("plugin sync", flag.ContinueOnError)
	moduleDir := fs.String("dir", ".", "project module directory")
	manifest := fs.String("manifest", "plugins.yaml", "project plugin manifest")
	out := fs.String("out", "internal/generated/fba_plugins.gen.go", "generated registration output")
	lockOut := fs.String("lock-out", "plugins.lock", "module-aware plugin lock output")
	packageName := fs.String("package", "generated", "generated package name")
	check := fs.Bool("check", false, "verify generated outputs without writing")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf(pluginSyncUsage)
	}
	if err := fbplugin.Sync(fbplugin.SyncOptions{
		ModuleDir: *moduleDir,
		Manifest:  *manifest,
		Out:       *out,
		LockOut:   *lockOut,
		Package:   *packageName,
		Check:     *check,
	}); err != nil {
		return err
	}
	if *check {
		_, err := fmt.Fprintln(stdout, "plugin state is synchronized; dependency updates were not checked")
		return err
	}
	return nil
}

func runPluginOutdated(args []string) error {
	fs := flag.NewFlagSet("plugin outdated", flag.ContinueOnError)
	moduleDir := fs.String("dir", ".", "project module directory")
	manifest := fs.String("manifest", "plugins.yaml", "project plugin manifest")
	if err := fs.Parse(args); err != nil {
		return err
	}
	statuses, err := fbplugin.Outdated(fbplugin.VersionOptions{
		ModuleDir: *moduleDir,
		Manifest:  *manifest,
		Targets:   fs.Args(),
	})
	if err != nil {
		return err
	}
	if len(statuses) == 0 {
		_, err := fmt.Fprintln(stdout, "no plugin modules found")
		return err
	}
	return writePluginModuleStatuses(statuses)
}

func runPluginUpdate(args []string) error {
	fs := flag.NewFlagSet("plugin update", flag.ContinueOnError)
	moduleDir := fs.String("dir", ".", "project module directory")
	manifest := fs.String("manifest", "plugins.yaml", "project plugin manifest")
	out := fs.String("out", "internal/generated/fba_plugins.gen.go", "generated registration output")
	lockOut := fs.String("lock-out", "plugins.lock", "module-aware plugin lock output")
	packageName := fs.String("package", "generated", "generated package name")
	version := fs.String("to", "", "target version for one plugin module")
	dryRun := fs.Bool("dry-run", false, "print updates without changing the project")
	if err := fs.Parse(args); err != nil {
		return err
	}
	updates, err := fbplugin.Update(fbplugin.UpdateOptions{
		ModuleDir: *moduleDir,
		Manifest:  *manifest,
		Out:       *out,
		LockOut:   *lockOut,
		Package:   *packageName,
		Targets:   fs.Args(),
		To:        *version,
		DryRun:    *dryRun,
	})
	if err != nil {
		return err
	}
	return writePluginUpdates(updates, *dryRun)
}

func writePluginModuleStatuses(statuses []fbplugin.ModuleStatus) error {
	writer := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "PLUGINS\tMODULE\tCURRENT\tAVAILABLE\tSTATUS"); err != nil {
		return err
	}
	for _, status := range statuses {
		current := displayedVersion(status.Current)
		available := displayedVersion(status.Available)
		state := "current"
		switch {
		case status.Main:
			current = "main"
			available = "-"
			state = "main"
		case status.Replace != nil:
			available = "-"
			state = "replace=" + displayedModuleReplace(status.Replace)
		case status.Available != "" && status.Available != status.Current:
			state = "update"
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\n",
			strings.Join(status.PluginIDs, ","),
			status.Module,
			current,
			available,
			state,
		); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func writePluginUpdates(updates []fbplugin.ModuleUpdate, dryRun bool) error {
	if len(updates) == 0 {
		_, err := fmt.Fprintln(stdout, "all selected plugin modules are current")
		return err
	}
	action := "updated"
	if dryRun {
		action = "would update"
	}
	for _, update := range updates {
		if _, err := fmt.Fprintf(
			stdout,
			"%s %s %s -> %s (plugins: %s)\n",
			action,
			update.Module,
			displayedVersion(update.From),
			displayedVersion(update.To),
			strings.Join(update.PluginIDs, ","),
		); err != nil {
			return err
		}
	}
	return nil
}

func displayedVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "-"
	}
	return version
}

func displayedModuleReplace(replacement *fbplugin.ModuleReplace) string {
	if replacement == nil {
		return "-"
	}
	if replacement.Version != "" {
		return replacement.Path + "@" + replacement.Version
	}
	return replacement.Path
}

func runPluginScan(args []string) error {
	fs := flag.NewFlagSet("plugin scan", flag.ContinueOnError)
	mode := fs.String("mode", "manifest", "comma-separated scan modes")
	moduleDir := fs.String("module", ".", "module directory")
	pluginsDir := fs.String("plugins-dir", "", "local plugins directory")
	manifest := fs.String("manifest", "", "plugins manifest")
	out := fs.String("out", "internal/generated/fba_plugins.gen.go", "generated registration output")
	lockOut := fs.String("lock-out", "", "plugin lock output")
	packageName := fs.String("package", "generated", "generated package name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	result, err := fbplugin.Scan(fbplugin.ScanOptions{
		Modes:      splitModes(*mode),
		ModuleDir:  *moduleDir,
		PluginsDir: *pluginsDir,
		Manifest:   *manifest,
	})
	if err != nil {
		return err
	}
	lockPath := *lockOut
	if lockPath == "" {
		lockPath = filepath.Join(filepath.Dir(*out), "plugin_manifest.lock")
	}
	if err := fbplugin.WriteLock(lockPath, result); err != nil {
		return err
	}
	return fbplugin.GenerateRegistration(*out, *packageName, result)
}

func runSwaggerScan(args []string) error {
	fs := flag.NewFlagSet("swagger scan", flag.ContinueOnError)
	plugins := fs.String("plugins", "internal/generated/plugin_manifest.lock", "plugin manifest lock")
	out := fs.String("out", "docs/openapi.json", "openapi output")
	title := fs.String("title", "FBA API", "document title")
	version := fs.String("version", "0.1.0", "document version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return fbswagger.Scan(fbswagger.ScanOptions{
		PluginLock: *plugins,
		Out:        *out,
		Title:      *title,
		Version:    *version,
	})
}

func splitModes(value string) []string {
	parts := strings.Split(value, ",")
	modes := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			modes = append(modes, part)
		}
	}
	return modes
}

func runContractSnapshot(args []string) error {
	fs := flag.NewFlagSet("contract snapshot", flag.ContinueOnError)
	contractDir := fs.String("contract", "contracts", "contract directory")
	out := fs.String("out", "internal/generated/api.contract.snapshot.json", "snapshot output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	contracts, err := fbcontract.Load(*contractDir)
	if err != nil {
		return err
	}
	snapshot, err := fbcontract.Snapshot(contracts)
	if err != nil {
		return err
	}
	return fbcontract.WriteSnapshot(*out, snapshot)
}

func runContractTest(args []string) error {
	fs := flag.NewFlagSet("contract test", flag.ContinueOnError)
	baseURL := fs.String("base-url", "", "base URL")
	contractDir := fs.String("contract", "contracts", "contract directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	contracts, err := fbcontract.Load(*contractDir)
	if err != nil {
		return err
	}
	result, err := fbcontract.Test(fbcontract.TestOptions{
		BaseURL:   *baseURL,
		Contracts: contracts,
	})
	if err != nil {
		return err
	}
	if !result.Passed {
		return fmt.Errorf("%s", fbcontract.FormatFailures(result))
	}
	return nil
}
