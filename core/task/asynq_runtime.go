package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const defaultWorkerShutdownTimeout = 8 * time.Second

// BackendRedisClient is the task engine's Redis connection. It is a distinct
// DI contract because task queues may use a different Redis database or
// cluster from application caching and realtime delivery.
type BackendRedisClient interface {
	redis.UniversalClient
}

// AsynqRuntimeOptions controls local worker processing. Queue weights follow
// Asynq's weighted-priority semantics; an empty map consumes the default queue.
type AsynqRuntimeOptions struct {
	Concurrency     int
	Queues          map[string]int
	ShutdownTimeout time.Duration
}

// ExecutionPayload is the stable envelope emitted by Runtime.Execute.
type ExecutionPayload struct {
	Args   any `json:"args"`
	Kwargs any `json:"kwargs"`
}

// AsynqRuntime owns the queue client, inspector, worker, and their shared Redis
// connection. Definitions must be registered before Start is called.
type AsynqRuntime struct {
	redis     redis.UniversalClient
	registry  DefinitionRegistry
	client    *asynq.Client
	inspector *asynq.Inspector
	server    *asynq.Server

	shutdownOnce sync.Once
	shutdownErr  error
}

func NewAsynqRuntime(redisClient redis.UniversalClient, registry DefinitionRegistry, opts AsynqRuntimeOptions) (*AsynqRuntime, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("task Redis client is required")
	}
	if registry == nil {
		return nil, fmt.Errorf("task definition registry is required")
	}
	shutdownTimeout := opts.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = defaultWorkerShutdownTimeout
	}
	queues := clonePositiveQueues(opts.Queues)
	return &AsynqRuntime{
		redis:     redisClient,
		registry:  registry,
		client:    asynq.NewClientFromRedisClient(redisClient),
		inspector: asynq.NewInspectorFromRedisClient(redisClient),
		server: asynq.NewServerFromRedisClient(redisClient, asynq.Config{
			Concurrency:     opts.Concurrency,
			Queues:          queues,
			ShutdownTimeout: shutdownTimeout,
		}),
	}, nil
}

// Start verifies Redis before reporting startup success. Asynq's Start method
// itself is non-blocking and otherwise would let an unavailable worker pass the
// application's start hooks unnoticed.
func (r *AsynqRuntime) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.server.Ping(); err != nil {
		return fmt.Errorf("start task worker: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.server.Start(BuildServeMux(r.registry)); err != nil {
		return fmt.Errorf("start task worker: %w", err)
	}
	return nil
}

// Shutdown is bounded by the Asynq server's ShutdownTimeout, then closes the
// shared Redis connection exactly once. Cleanup still runs when ctx is already
// canceled because abandoning the shared connection would leak resources.
func (r *AsynqRuntime) Shutdown(_ context.Context) error {
	r.shutdownOnce.Do(func() {
		r.server.Shutdown()
		r.shutdownErr = r.redis.Close()
	})
	return r.shutdownErr
}

func (r *AsynqRuntime) Reload(ctx context.Context) error {
	return ctx.Err()
}

func (r *AsynqRuntime) Execute(ctx context.Context, taskType string, args any, kwargs any) error {
	taskType = strings.TrimSpace(taskType)
	if taskType == "" {
		return fmt.Errorf("task is required")
	}
	definition, ok := findDefinition(r.registry, taskType)
	if !ok || definition.Handler == nil {
		return fmt.Errorf("task %q is not registered", taskType)
	}
	payload, err := json.Marshal(ExecutionPayload{Args: args, Kwargs: kwargs})
	if err != nil {
		return fmt.Errorf("encode task %q payload: %w", taskType, err)
	}
	options := make([]asynq.Option, 0, 1)
	if queue := strings.TrimSpace(definition.Queue); queue != "" {
		options = append(options, asynq.Queue(queue))
	}
	if _, err := r.client.EnqueueContext(ctx, asynq.NewTask(taskType, payload), options...); err != nil {
		return fmt.Errorf("enqueue task %q: %w", taskType, err)
	}
	return nil
}

func (r *AsynqRuntime) Cancel(ctx context.Context, taskID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return fmt.Errorf("task_id is required")
	}
	if err := r.inspector.CancelProcessing(taskID); err != nil {
		return fmt.Errorf("cancel task %q: %w", taskID, err)
	}
	return nil
}

func clonePositiveQueues(queues map[string]int) map[string]int {
	if len(queues) == 0 {
		return nil
	}
	cloned := make(map[string]int, len(queues))
	for name, priority := range queues {
		name = strings.TrimSpace(name)
		if name != "" && priority > 0 {
			cloned[name] = priority
		}
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}

func findDefinition(registry DefinitionRegistry, taskType string) (Definition, bool) {
	type lookupRegistry interface {
		Lookup(string) (Definition, bool)
	}
	if lookup, ok := registry.(lookupRegistry); ok {
		return lookup.Lookup(taskType)
	}
	for _, definition := range registry.All() {
		if definition.Type == taskType {
			return definition, true
		}
	}
	return Definition{}, false
}

var _ Runtime = (*AsynqRuntime)(nil)
