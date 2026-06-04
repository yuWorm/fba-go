package task

import (
	"context"
	"fmt"
)

// Runtime is the stable task execution contract consumed by business modules
// and task-management plugins. Concrete queue engines such as Asynq, Temporal,
// or a project-owned runtime should implement this interface instead of leaking
// their driver-specific APIs into application code.
type Runtime interface {
	Reload(context.Context) error
	Execute(ctx context.Context, task string, args any, kwargs any) error
	Cancel(ctx context.Context, taskID string) error
}

type NoopRuntime struct{}

func (NoopRuntime) Reload(context.Context) error {
	return nil
}

func (NoopRuntime) Execute(_ context.Context, task string, _ any, _ any) error {
	if task == "" {
		return fmt.Errorf("task is required")
	}
	return nil
}

func (NoopRuntime) Cancel(_ context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task_id is required")
	}
	return nil
}
