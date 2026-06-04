package plugin

import (
	"os"

	coreplugin "github.com/yuWorm/fba-go/core/plugin"
	"gopkg.in/yaml.v3"
)

type Plugin struct {
	ID      string          `json:"id" yaml:"id"`
	Name    string          `json:"name,omitempty" yaml:"name"`
	Module  string          `json:"module" yaml:"module"`
	Mode    coreplugin.Mode `json:"mode" yaml:"mode"`
	Swagger string          `json:"swagger,omitempty" yaml:"swagger"`
}

type Manifest struct {
	Plugins []Plugin `json:"plugins" yaml:"plugins"`
}

func ReadManifest(path string) (Manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}
