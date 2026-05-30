package swagger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	coreswagger "github.com/yuWorm/fba-go/core/swagger"
)

type ScanOptions struct {
	PluginLock string
	Out        string
	Title      string
	Version    string
}

type pluginLock struct {
	Plugins []pluginEntry `json:"plugins"`
}

type pluginEntry struct {
	ID      string `json:"id"`
	Swagger string `json:"swagger"`
}

func Scan(opts ScanOptions) error {
	if opts.PluginLock == "" {
		return fmt.Errorf("plugin lock path is required")
	}
	if opts.Out == "" {
		return fmt.Errorf("output path is required")
	}

	lock, err := readLock(opts.PluginLock)
	if err != nil {
		return err
	}

	fragments := make([]coreswagger.Fragment, 0, len(lock.Plugins))
	for _, plugin := range lock.Plugins {
		if plugin.Swagger == "" {
			continue
		}
		fragment, err := readFragment(plugin)
		if err != nil {
			return err
		}
		fragments = append(fragments, fragment)
	}

	doc, err := coreswagger.Aggregate(coreswagger.DocumentInfo{
		Title:   opts.Title,
		Version: opts.Version,
	}, fragments)
	if err != nil {
		return err
	}

	content, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(opts.Out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(opts.Out, content, 0o644)
}

func readLock(path string) (pluginLock, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return pluginLock{}, err
	}
	var lock pluginLock
	if err := json.Unmarshal(content, &lock); err != nil {
		return pluginLock{}, err
	}
	return lock, nil
}

func readFragment(plugin pluginEntry) (coreswagger.Fragment, error) {
	content, err := os.ReadFile(plugin.Swagger)
	if err != nil {
		return coreswagger.Fragment{}, err
	}
	var fragment coreswagger.Fragment
	if err := json.Unmarshal(content, &fragment); err != nil {
		return coreswagger.Fragment{}, err
	}
	if fragment.PluginID == "" {
		fragment.PluginID = plugin.ID
	}
	return fragment, nil
}
