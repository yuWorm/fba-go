package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	fbcontract "github.com/yuWorm/fba-go/cmd/fbago/internal/contract"
	fbplugin "github.com/yuWorm/fba-go/cmd/fbago/internal/plugin"
	"github.com/yuWorm/fba-go/cmd/fbago/internal/scaffold"
	fbswagger "github.com/yuWorm/fba-go/cmd/fbago/internal/swagger"
)

var stdout io.Writer = os.Stdout

const initUsage = "usage: fbago init <module> [--template TEMPLATE] [--dir DIR] [--core-replace PATH] [--core-version VERSION|latest]"

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
	case "swagger scan":
		return runSwaggerScan(args[2:])
	case "contract snapshot":
		return runContractSnapshot(args[2:])
	case "contract test":
		return runContractTest(args[2:])
	case "template list":
		return runTemplateList(args[2:])
	default:
		return fmt.Errorf("unknown command %s %s", args[0], args[1])
	}
}

func usage() error {
	return fmt.Errorf("usage: fbago <init|template|plugin|swagger|contract> [command]")
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
