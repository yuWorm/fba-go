package task_test

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/yuWorm/fba-go/core/task"
)

func TestRegistryRejectsDuplicateTaskTypes(t *testing.T) {
	registry := task.NewRegistry()
	definition := task.Definition{
		Type:    "email:send",
		Name:    "发送邮件",
		Queue:   "default",
		Handler: asynq.HandlerFunc(func(context.Context, *asynq.Task) error { return nil }),
	}

	if err := registry.Add(definition); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := registry.Add(definition); err == nil {
		t.Fatal("Add() duplicate error = nil, want error")
	}
}

func TestRegistryKeepsDefinitionsInRegistrationOrder(t *testing.T) {
	registry := task.NewRegistry()
	_ = registry.Add(task.Definition{Type: "critical:one", Name: "One", Queue: "critical"})
	_ = registry.Add(task.Definition{Type: "default:two", Name: "Two", Queue: "default"})

	definitions := registry.All()
	if len(definitions) != 2 {
		t.Fatalf("definitions = %d, want 2", len(definitions))
	}
	if definitions[0].Type != "critical:one" || definitions[1].Type != "default:two" {
		t.Fatalf("definition order = %+v", definitions)
	}
}

func TestRegistryImplementsDefinitionRegistryContract(t *testing.T) {
	var registry task.DefinitionRegistry = task.NewRegistry()
	_ = registry.Add(task.Definition{Type: "email:send", Name: "发送邮件"})

	definitions := registry.All()
	if len(definitions) != 1 || definitions[0].Type != "email:send" {
		t.Fatalf("definitions = %+v, want email:send", definitions)
	}
}

func TestNoopRuntimeValidatesRequiredTaskInputs(t *testing.T) {
	runtime := task.NoopRuntime{}

	if err := runtime.Execute(context.Background(), "", nil, nil); err == nil {
		t.Fatal("Execute() empty task error = nil, want error")
	}
	if err := runtime.Cancel(context.Background(), ""); err == nil {
		t.Fatal("Cancel() empty taskID error = nil, want error")
	}
	if err := runtime.Reload(context.Background()); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}
}

func TestMapAsynqStateReturnsPythonCompatibleStatus(t *testing.T) {
	cases := map[string]task.CompatibleStatus{
		"active":    task.StatusStarted,
		"completed": task.StatusSuccess,
		"retry":     task.StatusRetry,
		"archived":  task.StatusFailure,
		"pending":   task.StatusPending,
		"STARTED":   task.StatusStarted,
		"SUCCESS":   task.StatusSuccess,
		"RETRY":     task.StatusRetry,
		"FAILURE":   task.StatusFailure,
		"PENDING":   task.StatusPending,
	}

	for state, want := range cases {
		if got := task.MapAsynqState(state); got != want {
			t.Fatalf("MapAsynqState(%q) = %q, want %q", state, got, want)
		}
	}
}
