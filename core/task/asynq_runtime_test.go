package task_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	coretask "github.com/yuWorm/fba-go/core/task"
)

func TestAsynqRuntimeStartsWorkerAndExecutesRegisteredTask(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{redisServer.Addr()}})
	registry := coretask.NewRegistry()
	processed := make(chan coretask.ExecutionPayload, 1)
	if err := registry.Add(coretask.Definition{
		Type:  "fixture:execute",
		Name:  "Fixture execute",
		Queue: "critical",
		Handler: asynq.HandlerFunc(func(_ context.Context, task *asynq.Task) error {
			var payload coretask.ExecutionPayload
			if err := json.Unmarshal(task.Payload(), &payload); err != nil {
				return err
			}
			processed <- payload
			return nil
		}),
	}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	runtime, err := coretask.NewAsynqRuntime(redisClient, registry, coretask.AsynqRuntimeOptions{
		Concurrency: 1,
		Queues:      map[string]int{"critical": 1},
	})
	if err != nil {
		t.Fatalf("NewAsynqRuntime() error = %v", err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		if err := runtime.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown() error = %v", err)
		}
	})

	if err := runtime.Execute(context.Background(), "fixture:execute", []string{"hello"}, map[string]any{"count": 2}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	select {
	case payload := <-processed:
		args, ok := payload.Args.([]any)
		if !ok || len(args) != 1 || args[0] != "hello" {
			t.Fatalf("payload.Args = %#v, want [hello]", payload.Args)
		}
		kwargs, ok := payload.Kwargs.(map[string]any)
		if !ok || kwargs["count"] != float64(2) {
			t.Fatalf("payload.Kwargs = %#v, want count=2", payload.Kwargs)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("registered task was not processed by the worker")
	}
}

func TestAsynqRuntimeRejectsUnregisteredTask(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{redisServer.Addr()}})
	runtime, err := coretask.NewAsynqRuntime(redisClient, coretask.NewRegistry(), coretask.AsynqRuntimeOptions{})
	if err != nil {
		t.Fatalf("NewAsynqRuntime() error = %v", err)
	}
	defer runtime.Shutdown(context.Background())

	if err := runtime.Execute(context.Background(), "missing", nil, nil); err == nil {
		t.Fatal("Execute() error = nil, want unregistered-task error")
	}
}
