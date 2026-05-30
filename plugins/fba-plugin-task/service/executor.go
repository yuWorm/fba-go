package service

import (
	"context"
	"fmt"
)

type Executor interface {
	Reload(context.Context) error
	Execute(ctx context.Context, task string, args any, kwargs any) error
	Cancel(ctx context.Context, taskID string) error
}

type NoopExecutor struct{}

func (NoopExecutor) Reload(context.Context) error {
	return nil
}

func (NoopExecutor) Execute(_ context.Context, task string, _ any, _ any) error {
	if task == "" {
		return fmt.Errorf("task is required")
	}
	return nil
}

func (NoopExecutor) Cancel(_ context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task_id is required")
	}
	return nil
}
