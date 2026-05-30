package api

import "github.com/yuWorm/fba-go/core/plugin"

func Routes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/tasks/registered", "Registered tasks", h.RegisteredTasks, plugin.Auth()),
		plugin.DELETE("/tasks/:task_id/cancel", "Cancel task", h.CancelTask, plugin.Auth(), plugin.Perm("sys:task:revoke")),
		plugin.GET("/task-results/:pk", "Task result detail", h.GetTaskResult, plugin.Auth()),
		plugin.GET("/task-results", "List task results", h.ListTaskResults, plugin.Auth()),
		plugin.DELETE("/task-results", "Delete task results", h.DeleteTaskResults, plugin.Auth(), plugin.Perm("sys:task:del")),
		plugin.GET("/schedulers/all", "All schedulers", h.AllSchedulers, plugin.Auth()),
		plugin.GET("/schedulers/:pk", "Scheduler detail", h.GetScheduler, plugin.Auth()),
		plugin.GET("/schedulers", "List schedulers", h.ListSchedulers, plugin.Auth()),
		plugin.POST("/schedulers", "Create scheduler", h.CreateScheduler, plugin.Auth(), plugin.Perm("sys:task:add")),
		plugin.PUT("/schedulers/:pk", "Update scheduler", h.UpdateScheduler, plugin.Auth(), plugin.Perm("sys:task:edit")),
		plugin.PUT("/schedulers/:pk/status", "Update scheduler status", h.UpdateSchedulerStatus, plugin.Auth(), plugin.Perm("sys:task:edit")),
		plugin.DELETE("/schedulers/:pk", "Delete scheduler", h.DeleteScheduler, plugin.Auth(), plugin.Perm("sys:task:del")),
		plugin.POST("/schedulers/:pk/execute", "Execute scheduler", h.ExecuteScheduler, plugin.Auth(), plugin.Perm("sys:task:exec")),
	}
}
