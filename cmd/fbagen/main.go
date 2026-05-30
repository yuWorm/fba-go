package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	fbplugin "github.com/yuWorm/fba-go/cmd/fbagen/internal/plugin"
	fbswagger "github.com/yuWorm/fba-go/cmd/fbagen/internal/swagger"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: fbagen <plugin|swagger> <command>")
	}
	switch args[0] + " " + args[1] {
	case "plugin scan":
		return runPluginScan(args[2:])
	case "swagger scan":
		return runSwaggerScan(args[2:])
	default:
		return fmt.Errorf("unknown command %s %s", args[0], args[1])
	}
}

func runPluginScan(args []string) error {
	fs := flag.NewFlagSet("plugin scan", flag.ContinueOnError)
	mode := fs.String("mode", "manifest", "comma-separated scan modes")
	moduleDir := fs.String("module", ".", "module directory")
	pluginsDir := fs.String("plugins-dir", "", "local plugins directory")
	manifest := fs.String("manifest", "", "plugins manifest")
	out := fs.String("out", "internal/generated/fba_plugins.gen.go", "generated registration output")
	lockOut := fs.String("lock-out", "internal/generated/plugin_manifest.lock", "plugin lock output")
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
	if err := fbplugin.WriteLock(*lockOut, result); err != nil {
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
