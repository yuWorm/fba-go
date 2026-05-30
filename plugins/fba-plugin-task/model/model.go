package model

import "time"

type TaskScheduler struct {
	ID             int        `gorm:"column:id;primaryKey"`
	Name           string     `gorm:"column:name;size:64;uniqueIndex"`
	Task           string     `gorm:"column:task;size:256"`
	Args           *string    `gorm:"column:args;type:json"`
	Kwargs         *string    `gorm:"column:kwargs;type:json"`
	Queue          *string    `gorm:"column:queue;size:256"`
	Exchange       *string    `gorm:"column:exchange;size:256"`
	RoutingKey     *string    `gorm:"column:routing_key;size:256"`
	StartTime      *time.Time `gorm:"column:start_time"`
	ExpireTime     *time.Time `gorm:"column:expire_time"`
	ExpireSeconds  *int       `gorm:"column:expire_seconds"`
	Type           int        `gorm:"column:type"`
	IntervalEvery  *int       `gorm:"column:interval_every"`
	IntervalPeriod *string    `gorm:"column:interval_period;size:256"`
	Crontab        string     `gorm:"column:crontab;size:64"`
	OneOff         bool       `gorm:"column:one_off"`
	Enabled        bool       `gorm:"column:enabled"`
	TotalRunCount  int        `gorm:"column:total_run_count"`
	LastRunTime    *time.Time `gorm:"column:last_run_time"`
	Remark         *string    `gorm:"column:remark"`
	CreatedTime    time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime    *time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

func (TaskScheduler) TableName() string {
	return "task_scheduler"
}

type TaskResult struct {
	ID        int        `gorm:"column:id;primaryKey"`
	TaskID    string     `gorm:"column:task_id;size:155;uniqueIndex"`
	Status    string     `gorm:"column:status;size:64"`
	Result    *string    `gorm:"column:result;type:json"`
	DateDone  *time.Time `gorm:"column:date_done"`
	Traceback *string    `gorm:"column:traceback"`
	Name      *string    `gorm:"column:name;size:155"`
	Args      []byte     `gorm:"column:args"`
	Kwargs    []byte     `gorm:"column:kwargs"`
	Worker    *string    `gorm:"column:worker;size:155"`
	Retries   *int       `gorm:"column:retries"`
	Queue     *string    `gorm:"column:queue;size:155"`
}

func (TaskResult) TableName() string {
	return "task_result"
}

func SeedSchedulers() []TaskScheduler {
	interval := 10
	period := "seconds"
	return []TaskScheduler{
		{
			ID:             1,
			Name:           "Fixture",
			Task:           "task_demo",
			Type:           0,
			IntervalEvery:  &interval,
			IntervalPeriod: &period,
			Crontab:        "* * * * *",
			OneOff:         false,
			Enabled:        true,
			TotalRunCount:  0,
			CreatedTime:    seedTime(),
		},
	}
}

func SeedTaskResults() []TaskResult {
	name := "task_demo"
	worker := "worker-1"
	retries := 0
	queue := "default"
	done := seedTime()
	return []TaskResult{
		{
			ID:       1,
			TaskID:   "task-1",
			Status:   "active",
			DateDone: &done,
			Name:     &name,
			Args:     []byte("[]"),
			Kwargs:   []byte("{}"),
			Worker:   &worker,
			Retries:  &retries,
			Queue:    &queue,
		},
	}
}

func seedTime() time.Time {
	return time.Date(2026, 5, 30, 0, 0, 0, 0, time.Local)
}
