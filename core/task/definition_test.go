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
