package api

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
	"github.com/yuWorm/fba-plugin-task/dto"
	"github.com/yuWorm/fba-plugin-task/repo"
	"github.com/yuWorm/fba-plugin-task/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) Handler {
	if svc == nil {
		svc = service.New(repo.NewMemoryRepository(repo.SeedData()), nil, nil, nil)
	}
	return Handler{service: svc}
}

func (h Handler) RegisteredTasks(c fiber.Ctx) error {
	return c.JSON(response.Success(h.service.RegisteredTasks()))
}

func (h Handler) CancelTask(c fiber.Ctx) error {
	if err := h.service.CancelTask(c.RequestCtx(), c.Params("task_id")); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) GetTaskResult(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	item, err := h.service.GetTaskResult(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(item))
}

func (h Handler) ListTaskResults(c fiber.Ctx) error {
	page, size := pageParams(c)
	data, err := h.service.ListTaskResults(c.RequestCtx(), repo.ResultFilter{
		Name:   c.Query("name"),
		TaskID: c.Query("task_id"),
	}, page, size, "/api/v1/task-results")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(data))
}

func (h Handler) DeleteTaskResults(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.DeleteTaskResults(c.RequestCtx(), param.PKs); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) AllSchedulers(c fiber.Ctx) error {
	items, err := h.service.AllSchedulers(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(items))
}

func (h Handler) GetScheduler(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	item, err := h.service.GetScheduler(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(item))
}

func (h Handler) ListSchedulers(c fiber.Ctx) error {
	page, size := pageParams(c)
	data, err := h.service.ListSchedulers(c.RequestCtx(), repo.SchedulerFilter{
		Name: c.Query("name"),
		Type: intPtrQuery(c, "type"),
	}, page, size, "/api/v1/schedulers")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(data))
}

func (h Handler) CreateScheduler(c fiber.Ctx) error {
	var param dto.SchedulerParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.CreateScheduler(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateScheduler(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.SchedulerParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.UpdateScheduler(c.RequestCtx(), id, param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateSchedulerStatus(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	if err := h.service.ToggleSchedulerStatus(c.RequestCtx(), id); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteScheduler(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	if err := h.service.DeleteScheduler(c.RequestCtx(), id); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) ExecuteScheduler(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	if err := h.service.ExecuteScheduler(c.RequestCtx(), id); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func parseID(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	return id, nil
}

func pageParams(c fiber.Ctx) (int, int) {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(c.Query("size", "20"))
	if err != nil || size < 1 {
		size = 20
	}
	return page, size
}

func intPtrQuery(c fiber.Ctx, name string) *int {
	raw := c.Query(name)
	if raw == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &value
}
