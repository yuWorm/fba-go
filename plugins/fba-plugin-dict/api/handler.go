package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-go/core/response"
)

type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

type dictTypeDetail struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Remark      *string `json:"remark"`
	ID          int     `json:"id"`
	CreatedTime string  `json:"created_time"`
	UpdatedTime *string `json:"updated_time"`
}

type dictDataDetail struct {
	TypeID      int     `json:"type_id"`
	Label       string  `json:"label"`
	Value       string  `json:"value"`
	Color       *string `json:"color"`
	Sort        int     `json:"sort"`
	Status      int     `json:"status"`
	Remark      *string `json:"remark"`
	ID          int     `json:"id"`
	TypeCode    string  `json:"type_code"`
	CreatedTime string  `json:"created_time"`
	UpdatedTime *string `json:"updated_time"`
}

func (Handler) GetAllDictTypes(c fiber.Ctx) error {
	return c.JSON(response.Success(fixtureDictTypes()))
}

func (Handler) GetDictType(c fiber.Ctx) error {
	return c.JSON(response.Success(fixtureDictTypes()[0]))
}

func (Handler) ListDictTypes(c fiber.Ctx) error {
	items := fixtureDictTypes()
	return c.JSON(response.Success(pagination.NewPageData(items, int64(len(items)), 1, 20, "/api/v1/dict-types")))
}

func (Handler) CreateDictType(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdateDictType(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) DeleteDictTypes(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) GetAllDictData(c fiber.Ctx) error {
	return c.JSON(response.Success(fixtureDictData()))
}

func (Handler) GetDictData(c fiber.Ctx) error {
	return c.JSON(response.Success(fixtureDictData()[0]))
}

func (Handler) GetDictDataByTypeCode(c fiber.Ctx) error {
	code := c.Params("code")
	var items []dictDataDetail
	for _, item := range fixtureDictData() {
		if item.TypeCode == code {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []dictDataDetail{}
	}
	return c.JSON(response.Success(items))
}

func (Handler) ListDictData(c fiber.Ctx) error {
	items := fixtureDictData()
	return c.JSON(response.Success(pagination.NewPageData(items, int64(len(items)), 1, 20, "/api/v1/dict-datas")))
}

func (Handler) CreateDictData(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) UpdateDictData(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func (Handler) DeleteDictData(c fiber.Ctx) error {
	return c.JSON(response.Success[any](nil))
}

func fixtureDictTypes() []dictTypeDetail {
	sysStatusRemark := "系统通用状态：1/0"
	sysChooseRemark := "系统通用开关：true/false"
	return []dictTypeDetail{
		{
			ID:          1,
			Name:        "通用状态",
			Code:        "sys_status",
			Remark:      &sysStatusRemark,
			CreatedTime: "2026-05-30 00:00:00",
			UpdatedTime: nil,
		},
		{
			ID:          2,
			Name:        "通用开关",
			Code:        "sys_choose",
			Remark:      &sysChooseRemark,
			CreatedTime: "2026-05-30 00:00:00",
			UpdatedTime: nil,
		},
	}
}

func fixtureDictData() []dictDataDetail {
	disabledColor := "red"
	disabledRemark := "停用状态"
	enabledColor := "green"
	enabledRemark := "正常状态"
	closedColor := "error"
	closedRemark := "关闭状态"
	openColor := "success"
	openRemark := "开启状态"
	return []dictDataDetail{
		{
			ID:          1,
			TypeID:      1,
			TypeCode:    "sys_status",
			Label:       "停用",
			Value:       "0",
			Color:       &disabledColor,
			Sort:        1,
			Status:      1,
			Remark:      &disabledRemark,
			CreatedTime: "2026-05-30 00:00:00",
			UpdatedTime: nil,
		},
		{
			ID:          2,
			TypeID:      1,
			TypeCode:    "sys_status",
			Label:       "正常",
			Value:       "1",
			Color:       &enabledColor,
			Sort:        2,
			Status:      1,
			Remark:      &enabledRemark,
			CreatedTime: "2026-05-30 00:00:00",
			UpdatedTime: nil,
		},
		{
			ID:          3,
			TypeID:      2,
			TypeCode:    "sys_choose",
			Label:       "关闭",
			Value:       "false",
			Color:       &closedColor,
			Sort:        1,
			Status:      1,
			Remark:      &closedRemark,
			CreatedTime: "2026-05-30 00:00:00",
			UpdatedTime: nil,
		},
		{
			ID:          4,
			TypeID:      2,
			TypeCode:    "sys_choose",
			Label:       "开启",
			Value:       "true",
			Color:       &openColor,
			Sort:        2,
			Status:      1,
			Remark:      &openRemark,
			CreatedTime: "2026-05-30 00:00:00",
			UpdatedTime: nil,
		},
	}
}
