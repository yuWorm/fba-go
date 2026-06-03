package service_test

import (
	"context"
	"testing"

	"github.com/yuWorm/fba-go/core/realtime"
	coretask "github.com/yuWorm/fba-go/core/task"
	"github.com/yuWorm/fba-plugin-task/dto"
	"github.com/yuWorm/fba-plugin-task/repo"
	"github.com/yuWorm/fba-plugin-task/service"
)

func TestServiceRegisteredTasksUsesCoreRegistry(t *testing.T) {
	registry := coretask.NewRegistry()
	if err := registry.Add(coretask.Definition{Type: "task_demo", Name: "任务演示", Queue: "default"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), registry, &fakeExecutor{}, &fakeLeader{})

	items := svc.RegisteredTasks()

	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if items[0].Task != "task_demo" || items[0].Name != "任务演示" {
		t.Fatalf("registered task = %+v", items[0])
	}
}

func TestServiceSchedulerWritesReloadUnderLeaderLease(t *testing.T) {
	leader := &fakeLeader{acquire: true}
	executor := &fakeExecutor{}
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), nil, executor, leader)
	ctx := context.Background()

	if err := svc.CreateScheduler(ctx, schedulerParam()); err != nil {
		t.Fatalf("CreateScheduler() error = %v", err)
	}
	if err := svc.UpdateScheduler(ctx, 1, schedulerParam()); err != nil {
		t.Fatalf("UpdateScheduler() error = %v", err)
	}
	if err := svc.ToggleSchedulerStatus(ctx, 1); err != nil {
		t.Fatalf("ToggleSchedulerStatus() error = %v", err)
	}
	if err := svc.DeleteScheduler(ctx, 1); err != nil {
		t.Fatalf("DeleteScheduler() error = %v", err)
	}

	if leader.acquireCalls != 4 {
		t.Fatalf("Acquire() calls = %d, want 4", leader.acquireCalls)
	}
	if leader.releaseCalls != 4 {
		t.Fatalf("Release() calls = %d, want 4", leader.releaseCalls)
	}
	if executor.reloadCalls != 4 {
		t.Fatalf("Reload() calls = %d, want 4", executor.reloadCalls)
	}
}

func TestServiceDoesNotReloadWhenLeaderLeaseNotAcquired(t *testing.T) {
	leader := &fakeLeader{acquire: false}
	executor := &fakeExecutor{}
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), nil, executor, leader)

	if err := svc.CreateScheduler(context.Background(), schedulerParam()); err != nil {
		t.Fatalf("CreateScheduler() error = %v", err)
	}
	if executor.reloadCalls != 0 {
		t.Fatalf("Reload() calls = %d, want 0", executor.reloadCalls)
	}
}

func TestServiceExecuteSchedulerUsesExecutor(t *testing.T) {
	executor := &fakeExecutor{}
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), nil, executor, &fakeLeader{acquire: true})

	if err := svc.ExecuteScheduler(context.Background(), 1); err != nil {
		t.Fatalf("ExecuteScheduler() error = %v", err)
	}

	if executor.executedTask != "task_demo" {
		t.Fatalf("executed task = %q, want task_demo", executor.executedTask)
	}
}

func TestServiceExecuteSchedulerEmitsTaskNotifications(t *testing.T) {
	hub := realtime.NewMemoryHub(realtime.NewMemoryOnlineStore())
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), nil, &fakeExecutor{}, &fakeLeader{acquire: true}, service.WithRealtimeHub(hub))

	if err := svc.ExecuteScheduler(context.Background(), 1); err != nil {
		t.Fatalf("ExecuteScheduler() error = %v", err)
	}

	events := hub.Events()
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}
	if events[0].Event != "task_notification" || events[1].Event != "task_notification" {
		t.Fatalf("events = %+v, want task_notification events", events)
	}
	if got := events[0].Data.(realtime.TaskNotification).Msg; got != "任务 task_demo 开始执行" {
		t.Fatalf("first notification = %q", got)
	}
	if got := events[1].Data.(realtime.TaskNotification).Msg; got != "任务 task_demo 执行成功" {
		t.Fatalf("second notification = %q", got)
	}
}

func TestServiceMapsTaskResultStatus(t *testing.T) {
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), nil, &fakeExecutor{}, &fakeLeader{})

	result, err := svc.GetTaskResult(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTaskResult() error = %v", err)
	}
	if result.Status != "STARTED" {
		t.Fatalf("Status = %q, want STARTED", result.Status)
	}
}

func schedulerParam() dto.SchedulerParam {
	interval := 10
	period := "seconds"
	return dto.SchedulerParam{
		Name:           "Fixture",
		Task:           "task_demo",
		Type:           0,
		IntervalEvery:  &interval,
		IntervalPeriod: &period,
		Crontab:        "* * * * *",
		OneOff:         false,
	}
}

type fakeExecutor struct {
	reloadCalls  int
	executedTask string
}

func (f *fakeExecutor) Reload(context.Context) error {
	f.reloadCalls++
	return nil
}

func (f *fakeExecutor) Execute(_ context.Context, task string, _ any, _ any) error {
	f.executedTask = task
	return nil
}

func (f *fakeExecutor) Cancel(context.Context, string) error {
	return nil
}

type fakeLeader struct {
	acquire      bool
	acquireCalls int
	releaseCalls int
}

func (f *fakeLeader) Acquire(context.Context) (bool, error) {
	f.acquireCalls++
	return f.acquire, nil
}

func (f *fakeLeader) Release(context.Context) error {
	f.releaseCalls++
	return nil
}
