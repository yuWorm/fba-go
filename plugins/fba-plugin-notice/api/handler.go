package api

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
	"github.com/yuWorm/fba-plugin-notice/dto"
	"github.com/yuWorm/fba-plugin-notice/repo"
	"github.com/yuWorm/fba-plugin-notice/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) Handler {
	if svc == nil {
		svc = service.New(repo.NewMemoryRepository(repo.SeedData()))
	}
	return Handler{service: svc}
}

func (h Handler) GetNotice(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	item, err := h.service.Get(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(item))
}

func (h Handler) ListNotices(c fiber.Ctx) error {
	page, size := pageParams(c)
	data, err := h.service.List(c.RequestCtx(), repo.NoticeFilter{
		Title:  c.Query("title"),
		Type:   intPtrQuery(c, "type"),
		Status: intPtrQuery(c, "status"),
	}, page, size, "/api/v1/sys/notices")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(data))
}

func (h Handler) CreateNotice(c fiber.Ctx) error {
	var param dto.NoticeParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.Create(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateNotice(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.NoticeParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.Update(c.RequestCtx(), id, param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) DeleteNotices(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.Delete(c.RequestCtx(), param.PKs); err != nil {
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
