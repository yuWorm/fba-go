package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type Client interface {
	Enqueue(ctx context.Context, taskType string, payload any, opts ...asynq.Option) (*asynq.TaskInfo, error)
	Cancel(ctx context.Context, taskID string) error
}

type AsynqClient struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

func NewAsynqClient(client *asynq.Client, inspector *asynq.Inspector) *AsynqClient {
	return &AsynqClient{client: client, inspector: inspector}
}

func (c *AsynqClient) Enqueue(ctx context.Context, taskType string, payload any, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.client.EnqueueContext(ctx, asynq.NewTask(taskType, body), opts...)
}

func (c *AsynqClient) Cancel(ctx context.Context, taskID string) error {
	if c.inspector == nil {
		return fmt.Errorf("asynq inspector is not configured")
	}
	return c.inspector.CancelProcessing(taskID)
}
