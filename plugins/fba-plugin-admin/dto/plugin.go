package dto

import "github.com/yuWorm/fba-plugin-admin/model"

type PluginInstallParam struct {
	Type    string
	RepoURL string
	Name    string
}

type PluginConfigDetail struct {
	Plugin  PluginInfoDetail `json:"plugin"`
	App     map[string]any   `json:"app"`
	API     map[string]any   `json:"api,omitempty"`
	Setting map[string]any   `json:"settings"`
}

type PluginInfoDetail struct {
	Name        string   `json:"name"`
	Summary     string   `json:"summary"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	Database    []string `json:"database"`
	DependsOn   []string `json:"depends_on"`
	Enable      string   `json:"enable"`
}

func PluginFromModel(item model.Plugin) PluginConfigDetail {
	enable := "0"
	if item.Enabled {
		enable = "1"
	}
	return PluginConfigDetail{
		Plugin: PluginInfoDetail{
			Name:        item.ID,
			Summary:     item.Summary,
			Version:     item.Version,
			Description: item.Description,
			Author:      item.Author,
			Tags:        append([]string(nil), item.Tags...),
			Database:    append([]string(nil), item.Database...),
			DependsOn:   append([]string(nil), item.DependsOn...),
			Enable:      enable,
		},
		App: map[string]any{
			"extend": "admin",
		},
		Setting: map[string]any{},
	}
}

func PluginsFromModel(items []model.Plugin) []PluginConfigDetail {
	result := make([]PluginConfigDetail, 0, len(items))
	for _, item := range items {
		result = append(result, PluginFromModel(item))
	}
	return result
}
