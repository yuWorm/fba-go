package task

import (
	"github.com/yuWorm/fba-go/core/plugin"
	coretask "github.com/yuWorm/fba-go/core/task"
	taskapi "github.com/yuWorm/fba-plugin-task/api"
)

func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{
		ID:                "task",
		Name:              "Task Plugin",
		Version:           "0.1.0",
		Description:       "Task scheduler compatibility plugin",
		DependsOn:         []plugin.Dependency{{ID: "admin", Optional: true}},
		AutoInjectDefault: true,
	}
}

func (Module) Register(ctx plugin.Context) error {
	var registry *coretask.Registry
	_ = ctx.Container().Resolve(&registry)
	handler := taskapi.NewHandler(registry)

	for _, route := range []plugin.Route{
		{Method: "GET", Path: "/tasks/registered", Summary: "Registered tasks", AuthRequired: true, Handler: handler.RegisteredTasks},
		{Method: "DELETE", Path: "/tasks/:task_id/cancel", Summary: "Cancel task", AuthRequired: true, Permission: "sys:task:revoke", Handler: handler.CancelTask},
		{Method: "GET", Path: "/task-results/:pk", Summary: "Task result detail", AuthRequired: true, Handler: handler.GetTaskResult},
		{Method: "GET", Path: "/task-results", Summary: "List task results", AuthRequired: true, Handler: handler.ListTaskResults},
		{Method: "DELETE", Path: "/task-results", Summary: "Delete task results", AuthRequired: true, Permission: "sys:task:del", Handler: handler.DeleteTaskResults},
		{Method: "GET", Path: "/schedulers/all", Summary: "All schedulers", AuthRequired: true, Handler: handler.AllSchedulers},
		{Method: "GET", Path: "/schedulers/:pk", Summary: "Scheduler detail", AuthRequired: true, Handler: handler.GetScheduler},
		{Method: "GET", Path: "/schedulers", Summary: "List schedulers", AuthRequired: true, Handler: handler.ListSchedulers},
		{Method: "POST", Path: "/schedulers", Summary: "Create scheduler", AuthRequired: true, Permission: "sys:task:add", Handler: handler.CreateScheduler},
		{Method: "PUT", Path: "/schedulers/:pk", Summary: "Update scheduler", AuthRequired: true, Permission: "sys:task:edit", Handler: handler.UpdateScheduler},
		{Method: "PUT", Path: "/schedulers/:pk/status", Summary: "Update scheduler status", AuthRequired: true, Permission: "sys:task:edit", Handler: handler.UpdateSchedulerStatus},
		{Method: "DELETE", Path: "/schedulers/:pk", Summary: "Delete scheduler", AuthRequired: true, Permission: "sys:task:del", Handler: handler.DeleteScheduler},
		{Method: "POST", Path: "/schedulers/:pk/execute", Summary: "Execute scheduler", AuthRequired: true, Permission: "sys:task:exec", Handler: handler.ExecuteScheduler},
	} {
		if err := ctx.Route(route); err != nil {
			return err
		}
	}

	return nil
}
