package dto

import (
	"time"

	"github.com/yuWorm/fba-plugin-config/model"
)

const TimeLayout = "2006-01-02 15:04:05"

type ConfigDetail struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Type        *string `json:"type"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	IsFrontend  bool    `json:"is_frontend"`
	Remark      *string `json:"remark"`
	CreatedTime string  `json:"created_time"`
	UpdatedTime *string `json:"updated_time"`
}

type ConfigParam struct {
	Name       string  `json:"name"`
	Type       *string `json:"type"`
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	IsFrontend bool    `json:"is_frontend"`
	Remark     *string `json:"remark"`
}

type ConfigBulkParam struct {
	ID int `json:"id"`
	ConfigParam
}

func ConfigFromModel(item model.Config) ConfigDetail {
	return ConfigDetail{
		ID:          item.ID,
		Name:        item.Name,
		Type:        item.Type,
		Key:         item.Key,
		Value:       item.Value,
		IsFrontend:  item.IsFrontend,
		Remark:      item.Remark,
		CreatedTime: formatTime(item.CreatedTime),
		UpdatedTime: formatTimePtr(item.UpdatedTime),
	}
}

func ConfigListFromModel(items []model.Config) []ConfigDetail {
	result := make([]ConfigDetail, 0, len(items))
	for _, item := range items {
		result = append(result, ConfigFromModel(item))
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
