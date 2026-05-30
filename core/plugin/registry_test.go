package plugin_test

import (
	"strings"
	"testing"

	"github.com/yuWorm/fba-go/core/plugin"
)

func TestRegistrySortsDependenciesAndSkipsPureDependencyRegistration(t *testing.T) {
	reg := plugin.NewRegistry()
	admin := &registryModule{id: "admin"}
	sdk := &registryModule{id: "oauth2-sdk"}
	order := &registryModule{id: "order", deps: []plugin.Dependency{
		{ID: "admin"},
		{ID: "oauth2-sdk"},
	}}

	if err := reg.Add(order, plugin.ModeAuto); err != nil {
		t.Fatalf("Add(order) error = %v", err)
	}
	if err := reg.Add(sdk, plugin.ModePureDependency); err != nil {
		t.Fatalf("Add(sdk) error = %v", err)
	}
	if err := reg.Add(admin, plugin.ModeAuto); err != nil {
		t.Fatalf("Add(admin) error = %v", err)
	}

	ordered, err := reg.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	got := orderedIDs(ordered)
	if got != "admin,oauth2-sdk,order" {
		t.Fatalf("resolved order = %q", got)
	}

	if err := reg.RegisterAll(plugin.NewContext(plugin.ContextOptions{})); err != nil {
		t.Fatalf("RegisterAll() error = %v", err)
	}
	if !admin.registered {
		t.Fatal("admin registered = false, want true")
	}
	if sdk.registered {
		t.Fatal("pure dependency registered = true, want false")
	}
	if !order.registered {
		t.Fatal("order registered = false, want true")
	}
}

func TestRegistryRejectsMissingRequiredDependency(t *testing.T) {
	reg := plugin.NewRegistry()
	if err := reg.Add(&registryModule{id: "order", deps: []plugin.Dependency{{ID: "admin"}}}, plugin.ModeAuto); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err := reg.RegisterAll(plugin.NewContext(plugin.ContextOptions{}))
	if err == nil || !strings.Contains(err.Error(), "missing dependency") {
		t.Fatalf("RegisterAll() error = %v, want missing dependency", err)
	}
}

func TestRegistryRejectsCycles(t *testing.T) {
	reg := plugin.NewRegistry()
	_ = reg.Add(&registryModule{id: "order", deps: []plugin.Dependency{{ID: "payment"}}}, plugin.ModeAuto)
	_ = reg.Add(&registryModule{id: "payment", deps: []plugin.Dependency{{ID: "order"}}}, plugin.ModeAuto)

	err := reg.RegisterAll(plugin.NewContext(plugin.ContextOptions{}))
	if err == nil || !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("RegisterAll() error = %v, want dependency cycle", err)
	}
}

func TestRegistryRejectsDisabledRequiredDependency(t *testing.T) {
	reg := plugin.NewRegistry()
	_ = reg.Add(&registryModule{id: "admin"}, plugin.ModeDisabled)
	_ = reg.Add(&registryModule{id: "order", deps: []plugin.Dependency{{ID: "admin"}}}, plugin.ModeAuto)

	err := reg.RegisterAll(plugin.NewContext(plugin.ContextOptions{}))
	if err == nil || !strings.Contains(err.Error(), "disabled dependency") {
		t.Fatalf("RegisterAll() error = %v, want disabled dependency", err)
	}
}

type registryModule struct {
	id         string
	deps       []plugin.Dependency
	registered bool
}

func (m *registryModule) Meta() plugin.Meta {
	return plugin.Meta{ID: m.id, Version: "0.1.0", DependsOn: m.deps}
}

func (m *registryModule) Register(plugin.Context) error {
	m.registered = true
	return nil
}

func orderedIDs(entries []plugin.Entry) string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.Module.Meta().ID)
	}
	return strings.Join(ids, ",")
}
