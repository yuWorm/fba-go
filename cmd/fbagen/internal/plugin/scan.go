package plugin

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	coreplugin "github.com/yuWorm/fba-go/core/plugin"
	"gopkg.in/yaml.v3"
)

type ScanOptions struct {
	Modes      []string
	ModuleDir  string
	PluginsDir string
	Manifest   string
}

type ScanResult struct {
	Plugins []Plugin `json:"plugins"`
}

func Scan(opts ScanOptions) (ScanResult, error) {
	if len(opts.Modes) == 0 {
		opts.Modes = []string{"manifest"}
	}
	merged := map[string]Plugin{}

	for _, mode := range opts.Modes {
		switch mode {
		case "manifest":
			if opts.Manifest == "" {
				continue
			}
			manifest, err := ReadManifest(opts.Manifest)
			if err != nil {
				return ScanResult{}, err
			}
			mergePlugins(merged, manifest.Plugins)
		case "local":
			plugins, err := scanLocal(opts.PluginsDir)
			if err != nil {
				return ScanResult{}, err
			}
			mergePlugins(merged, plugins)
		case "imports":
			plugins, err := scanImports(opts.ModuleDir)
			if err != nil {
				return ScanResult{}, err
			}
			mergePlugins(merged, plugins)
		}
	}

	plugins := make([]Plugin, 0, len(merged))
	for _, plugin := range merged {
		if plugin.Mode == "" {
			plugin.Mode = coreplugin.ModeAuto
		}
		plugins = append(plugins, plugin)
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].ID < plugins[j].ID
	})
	return ScanResult{Plugins: plugins}, nil
}

func mergePlugins(dst map[string]Plugin, plugins []Plugin) {
	for _, plugin := range plugins {
		if plugin.ID == "" {
			plugin.ID = moduleID(plugin.Module)
		}
		if plugin.Mode == "" {
			plugin.Mode = coreplugin.ModeAuto
		}
		dst[plugin.ID] = plugin
	}
}

func scanLocal(pluginsDir string) ([]Plugin, error) {
	if pluginsDir == "" {
		return nil, nil
	}
	var plugins []Plugin
	err := filepath.WalkDir(pluginsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "plugin.yaml" {
			return nil
		}
		manifest, err := readPluginYAML(path)
		if err != nil {
			return err
		}
		plugins = append(plugins, manifest)
		return nil
	})
	return plugins, err
}

func readPluginYAML(path string) (Plugin, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Plugin{}, err
	}
	var plugin Plugin
	if err := yaml.Unmarshal(content, &plugin); err != nil {
		return Plugin{}, err
	}
	return plugin, nil
}

func scanImports(moduleDir string) ([]Plugin, error) {
	if moduleDir == "" {
		return nil, nil
	}
	var plugins []Plugin
	err := filepath.WalkDir(moduleDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		filePlugins, err := scanImportFile(path)
		if err != nil {
			return err
		}
		plugins = append(plugins, filePlugins...)
		return nil
	})
	return plugins, err
}

func scanImportFile(path string) ([]Plugin, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var plugins []Plugin
	for _, spec := range file.Imports {
		if spec.Name == nil || spec.Name.Name != "_" {
			continue
		}
		module := strings.Trim(spec.Path.Value, `"`)
		id := moduleID(module)
		if strings.HasPrefix(id, "fba-plugin-") {
			id = strings.TrimPrefix(id, "fba-plugin-")
		}
		plugins = append(plugins, Plugin{ID: id, Module: module, Mode: coreplugin.ModeAuto})
	}
	return plugins, nil
}

func moduleID(module string) string {
	parts := strings.Split(strings.Trim(module, "/"), "/")
	if len(parts) == 0 {
		return module
	}
	return parts[len(parts)-1]
}
