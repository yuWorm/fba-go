package api

import (
	stderrors "errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/fiberx"
	"github.com/yuWorm/fba-go/core/response"
	"github.com/yuWorm/fba-plugin-dict/dto"
	"github.com/yuWorm/fba-plugin-dict/repo"
	"github.com/yuWorm/fba-plugin-dict/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) Handler {
	if svc == nil {
		svc = service.New(repo.NewMemoryRepository(repo.SeedData()), service.NoopInvalidator{})
	}
	return Handler{service: svc}
}

func (h Handler) GetAllDictTypes(c fiber.Ctx) error {
	items, err := h.service.AllTypes(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(items))
}

func (h Handler) GetDictType(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	item, err := h.service.GetType(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(item))
}

func (h Handler) ListDictTypes(c fiber.Ctx) error {
	page, size := pageParams(c)
	data, err := h.service.ListTypes(c.RequestCtx(), repo.DictTypeFilter{
		Name: c.Query("name"),
		Code: c.Query("code"),
	}, page, size, "/api/v1/dict-types")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(data))
}

func (h Handler) CreateDictType(c fiber.Ctx) error {
	var param dto.DictTypeParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.CreateType(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateDictType(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.DictTypeParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	return mutationSuccess(c, h.service.UpdateType(c.RequestCtx(), id, param))
}

func (h Handler) DeleteDictTypes(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	return mutationSuccess(c, h.service.DeleteTypes(c.RequestCtx(), param.PKs))
}

func (h Handler) GetAllDictData(c fiber.Ctx) error {
	items, err := h.service.AllData(c.RequestCtx())
	if err != nil {
		return err
	}
	return c.JSON(response.Success(items))
}

func (h Handler) GetDictData(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	item, err := h.service.GetData(c.RequestCtx(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.Success(item))
}

func (h Handler) GetDictDataByTypeCode(c fiber.Ctx) error {
	items, err := h.service.GetDataByTypeCode(c.RequestCtx(), c.Params("code"))
	if err != nil {
		return err
	}
	return c.JSON(response.Success(items))
}

func (h Handler) ListDictData(c fiber.Ctx) error {
	page, size := pageParams(c)
	data, err := h.service.ListData(c.RequestCtx(), repo.DictDataFilter{
		TypeCode: c.Query("type_code"),
		Label:    c.Query("label"),
		Value:    c.Query("value"),
		Status:   intPtrQuery(c, "status"),
		TypeID:   intPtrQuery(c, "type_id"),
	}, page, size, "/api/v1/dict-datas")
	if err != nil {
		return err
	}
	return c.JSON(response.Success(data))
}

func (h Handler) CreateDictData(c fiber.Ctx) error {
	var param dto.DictDataParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.CreateData(c.RequestCtx(), param); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

func (h Handler) UpdateDictData(c fiber.Ctx) error {
	id, err := parseID(c.Params("pk"))
	if err != nil {
		return err
	}
	var param dto.DictDataParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	return mutationSuccess(c, h.service.UpdateData(c.RequestCtx(), id, param))
}

func (h Handler) DeleteDictData(c fiber.Ctx) error {
	var param dto.DeleteParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	return mutationSuccess(c, h.service.DeleteData(c.RequestCtx(), param.PKs))
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
