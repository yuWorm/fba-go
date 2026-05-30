package swagger

import (
	"fmt"
	"strings"
)

func Aggregate(info DocumentInfo, fragments []Fragment) (Document, error) {
	if info.Title == "" {
		info.Title = "FBA API"
	}
	if info.Version == "" {
		info.Version = "0.1.0"
	}

	doc := Document{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:   info.Title,
			Version: info.Version,
		},
		Paths: Paths{},
		Components: Components{
			Schemas: map[string]any{},
		},
	}

	for _, fragment := range fragments {
		if err := mergeFragment(&doc, fragment); err != nil {
			return Document{}, err
		}
	}

	return doc, nil
}

func mergeFragment(doc *Document, fragment Fragment) error {
	for path, item := range fragment.Paths {
		if doc.Paths[path] == nil {
			doc.Paths[path] = PathItem{}
		}
		for method, operation := range item {
			normalizedMethod := strings.ToLower(method)
			if _, exists := doc.Paths[path][normalizedMethod]; exists {
				return fmt.Errorf("duplicate route %s %s", strings.ToUpper(normalizedMethod), path)
			}
			doc.Paths[path][normalizedMethod] = operation
		}
	}

	for name, schema := range fragment.Schemas {
		key := name
		if fragment.PluginID != "" {
			key = fragment.PluginID + "." + name
		}
		if _, exists := doc.Components.Schemas[key]; exists {
			return fmt.Errorf("duplicate schema %s", key)
		}
		doc.Components.Schemas[key] = schema
	}

	return nil
}
