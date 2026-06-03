package api

import (
	stderrors "errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/fiberx"
	"github.com/yuWorm/fba-go/core/response"
	"github.com/yuWorm/fba-plugin-config/dto"
	"github.com/yuWorm/fba-plugin-config/repo"
	"github.com/yuWorm/fba-plugin-config/service"
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

func (h Handler) GetAllConfigs(c fiber.Ctx) error {
	items, err := h.service.All(c.RequestCtx(), c.Query("type"))
	if err != nil {
		return err
	}
	return c.JSON(response.Success(items))
}

func (h Handler) GetConfig(c fiber.Ctx) error {
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

func (h Handler) ListConfigs(c fiber.Ctx) error {
	page, size := pageParams(c)
	data, err := h.service.List(c.RequestCtx(), repo.ConfigFilter{
		Name: c.Query("name"),
		Type: c.Query("type"),
	}, page, size, "/api/v1/sys/configs")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(data))
}

func (h Handler) CreateConfig(c fiber.Ctx) error {
	var param dto.ConfigParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.Create(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) BulkUpdateConfigs(c fiber.Ctx) error {
	var params []dto.ConfigBulkParam
	if err := c.Bind().Body(&params); err != nil {
		return err
	}
	return mutationSuccess(c, h.service.BulkUpdate(c.RequestCtx(), params))
}

func (h Handler) UpdateConfig(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.ConfigParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	return mutationSuccess(c, h.service.Update(c.RequestCtx(), id, param))
}

func (h Handler) DeleteConfigs(c fiber.Ctx) error {
	var ids []int
	// The Python route receives a raw list[int] body instead of {"pks":[]};
	// keep that shape to avoid forcing frontend callers to branch by runtime.
	if err := c.Bind().Body(&ids); err != nil {
		return err
	}
	return mutationSuccess(c, h.service.Delete(c.RequestCtx(), ids))
}

func mutationSuccess(c fiber.Ctx, err error) error {
	if err == nil {
		return c.JSON(response.Success[any](nil))
	}
	if isRawRepoNotFound(err) {
		return c.JSON(response.Fail[any](nil))
	}
	return err
}

func isRawRepoNotFound(err error) bool {
	var appErr *fbaerrors.AppError
	// Wrapped service errors are Python-style 404 guards; only raw repo misses represent count == 0.
	if stderrors.As(err, &appErr) {
		return false
	}
	return stderrors.Is(err, repo.ErrNotFound)
}

func parseID(raw string) (int, error) {
	return fiberx.ParseIntParam("pk", raw)
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
