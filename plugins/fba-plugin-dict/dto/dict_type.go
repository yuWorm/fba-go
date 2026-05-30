package dto

import (
	"time"

	"github.com/yuWorm/fba-plugin-dict/model"
)

const TimeLayout = "2006-01-02 15:04:05"

type DictTypeDetail struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Remark      *string `json:"remark"`
	ID          int     `json:"id"`
	CreatedTime string  `json:"created_time"`
	UpdatedTime *string `json:"updated_time"`
}

type DictTypeParam struct {
	Name   string  `json:"name"`
	Code   string  `json:"code"`
	Remark *string `json:"remark"`
}

type DeleteParam struct {
	PKs []int `json:"pks"`
}

func DictTypeFromModel(item model.DictType) DictTypeDetail {
	return DictTypeDetail{
		ID:          item.ID,
		Name:        item.Name,
		Code:        item.Code,
		Remark:      item.Remark,
		CreatedTime: formatTime(item.CreatedTime),
		UpdatedTime: formatTimePtr(item.UpdatedTime),
	}
}

func DictTypesFromModel(items []model.DictType) []DictTypeDetail {
	result := make([]DictTypeDetail, 0, len(items))
	for _, item := range items {
		result = append(result, DictTypeFromModel(item))
	}
	return result
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(TimeLayout)
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatTime(*value)
	return &formatted
}
