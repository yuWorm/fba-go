package api

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-go/core/response"
	coretask "github.com/yuWorm/fba-go/core/task"
)

type Handler struct {
	registry *coretask.Registry
}

func NewHandler(registry *coretask.Registry) Handler {
	return Handler{registry: registry}
}

type registeredTask struct {
	Name string `json:"name"`
	Task string `json:"task"`
}

type schedulerDetail struct {
	Name           string  `json:"name"`
	Task           string  `json:"task"`
	Args           any     `json:"args"`
	Kwargs         any     `json:"kwargs"`
	Queue          *string `json:"queue"`
	Exchange       *string `json:"exchange"`
	RoutingKey     *string `json:"routing_key"`
	StartTime      *string `json:"start_time"`
	ExpireTime     *string `json:"expire_time"`
	ExpireSeconds  *int    `json:"expire_seconds"`
	Type           int     `json:"type"`
	IntervalEvery  *int    `json:"interval_every"`
	IntervalPeriod *string `json:"interval_period"`
	Crontab        string  `json:"crontab"`
	OneOff         bool    `json:"one_off"`
	Remark         *string `json:"remark"`
	ID             int     `json:"id"`
	Enabled        bool    `json:"enabled"`
	TotalRunCount  int     `json:"total_run_count"`
	LastRunTime    *string `json:"last_run_time"`
	CreatedTime    string  `json:"created_time"`
	UpdatedTime    *string `json:"updated_time"`
}

type taskResultDetail struct {
	TaskID    string  `json:"task_id"`
	Status    string  `json:"status"`
	Result    any     `json:"result"`
	DateDone  *string `json:"date_done"`
	Traceback *string `json:"traceback"`
	Name      *string `json:"name"`
	Args      any     `json:"args"`
	Kwargs    any     `json:"kwargs"`
	Worker    *string `json:"worker"`
	Retries   *int    `json:"retries"`
	Queue     *string `json:"queue"`
	ID        int     `json:"id"`
}

func (h Handler) RegisteredTasks(c fiber.Ctx) error {
	items := h.registered()
	return c.JSON(response.Success(items))
}

func (Handler) CancelTask(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) GetTaskResult(c fiber.Ctx) error {
	return c.JSON(response.Success(fixtureTaskResults()[0]))
}

func (Handler) ListTaskResults(c fiber.Ctx) error {
	items := fixtureTaskResults()
	return c.JSON(response.Success(pagination.NewPageData(items, int64(len(items)), page(c), size(c), "/api/v1/task-results")))
}

func (Handler) DeleteTaskResults(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) AllSchedulers(c fiber.Ctx) error {
	return c.JSON(response.Success(fixtureSchedulers()))
}

func (Handler) GetScheduler(c fiber.Ctx) error {
	return c.JSON(response.Success(fixtureSchedulers()[0]))
}

func (Handler) ListSchedulers(c fiber.Ctx) error {
	items := fixtureSchedulers()
	return c.JSON(response.Success(pagination.NewPageData(items, int64(len(items)), page(c), size(c), "/api/v1/schedulers")))
}

func (Handler) CreateScheduler(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdateScheduler(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdateSchedulerStatus(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) DeleteScheduler(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) ExecuteScheduler(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (h Handler) registered() []registeredTask {
	if h.registry == nil {
		return []registeredTask{{Name: "task_demo", Task: "task_demo"}}
	}
	definitions := h.registry.All()
	if len(definitions) == 0 {
		return []registeredTask{}
	}
	items := make([]registeredTask, 0, len(definitions))
	for _, definition := range definitions {
		name := definition.Name
		if name == "" {
			name = definition.Type
		}
		items = append(items, registeredTask{Name: name, Task: definition.Type})
	}
	return items
}

func fixtureSchedulers() []schedulerDetail {
	interval := 10
	period := "seconds"
	return []schedulerDetail{
		{
			ID:             1,
			Name:           "Fixture",
			Task:           "task_demo",
			Args:           nil,
			Kwargs:         nil,
			Queue:          nil,
			Exchange:       nil,
			RoutingKey:     nil,
			StartTime:      nil,
			ExpireTime:     nil,
			ExpireSeconds:  nil,
			Type:           0,
			IntervalEvery:  &interval,
			IntervalPeriod: &period,
			Crontab:        "* * * * *",
			OneOff:         false,
			Remark:         nil,
			Enabled:        true,
			TotalRunCount:  0,
			LastRunTime:    nil,
			CreatedTime:    "2026-05-30 00:00:00",
			UpdatedTime:    nil,
		},
	}
}

func fixtureTaskResults() []taskResultDetail {
	name := "task_demo"
	worker := "worker-1"
	retries := 0
	queue := "default"
	dateDone := "2026-05-30 00:00:00"
	return []taskResultDetail{
		{
			ID:        1,
			TaskID:    "task-1",
			Status:    string(coretask.MapAsynqState("active")),
			Result:    nil,
			DateDone:  &dateDone,
			Traceback: nil,
			Name:      &name,
			Args:      []any{},
			Kwargs:    map[string]any{},
			Worker:    &worker,
			Retries:   &retries,
			Queue:     &queue,
		},
	}
}

func page(c fiber.Ctx) int {
	value, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || value < 1 {
		return 1
	}
	return value
}

func size(c fiber.Ctx) int {
	value, err := strconv.Atoi(c.Query("size", "20"))
	if err != nil || value < 1 {
		return 20
	}
	return value
}
