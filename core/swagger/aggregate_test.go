package swagger_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yuWorm/fba-go/core/swagger"
)

func TestAggregateMergesPathsAndPrefixesSchemas(t *testing.T) {
	doc, err := swagger.Aggregate(swagger.DocumentInfo{
		Title:   "FBA API",
		Version: "0.1.0",
	}, []swagger.Fragment{
		{
			PluginID: "admin",
			Paths: map[string]swagger.PathItem{
				"/api/v1/sys/users": {
					"get": swagger.Operation{Summary: "list users"},
				},
			},
			Schemas: map[string]any{
				"User": map[string]any{"type": "object"},
			},
		},
		{
			PluginID: "dict",
			Paths: map[string]swagger.PathItem{
				"/api/v1/dict-datas": {
					"get": swagger.Operation{Summary: "list dict data"},
				},
			},
			Schemas: map[string]any{
				"Item": map[string]any{"type": "object"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	if doc.OpenAPI != "3.0.3" {
		t.Fatalf("OpenAPI = %q, want 3.0.3", doc.OpenAPI)
	}
	if doc.Paths["/api/v1/sys/users"]["get"].Summary != "list users" {
		t.Fatalf("admin path not merged: %+v", doc.Paths)
	}
	if _, ok := doc.Components.Schemas["admin.User"]; !ok {
		t.Fatalf("admin.User schema missing: %+v", doc.Components.Schemas)
	}
	if _, ok := doc.Components.Schemas["dict.Item"]; !ok {
		t.Fatalf("dict.Item schema missing: %+v", doc.Components.Schemas)
	}
}

func TestAggregateRejectsDuplicateMethodPath(t *testing.T) {
	_, err := swagger.Aggregate(swagger.DocumentInfo{}, []swagger.Fragment{
		{
			PluginID: "admin",
			Paths: map[string]swagger.PathItem{
				"/api/v1/shared": {"get": swagger.Operation{Summary: "admin"}},
			},
		},
		{
			PluginID: "dict",
			Paths: map[string]swagger.PathItem{
				"/api/v1/shared": {"GET": swagger.Operation{Summary: "dict"}},
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate route") {
		t.Fatalf("Aggregate() error = %v, want duplicate route", err)
	}
}

func TestDocumentMarshalsOpenAPIFields(t *testing.T) {
	doc, err := swagger.Aggregate(swagger.DocumentInfo{Title: "FBA", Version: "1.0.0"}, nil)
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}

	got, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !strings.Contains(string(got), `"openapi":"3.0.3"`) {
		t.Fatalf("document JSON = %s", got)
	}
}
