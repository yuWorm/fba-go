package dto

import "github.com/yuWorm/fba-plugin-dict/model"

type DictDataDetail struct {
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

type DictDataParam struct {
	TypeID int     `json:"type_id"`
	Label  string  `json:"label"`
	Value  string  `json:"value"`
	Color  *string `json:"color"`
	Sort   int     `json:"sort"`
	Status int     `json:"status"`
	Remark *string `json:"remark"`
}

func DictDataFromModel(item model.DictData) DictDataDetail {
	return DictDataDetail{
		ID:          item.ID,
		TypeID:      item.TypeID,
		TypeCode:    item.TypeCode,
		Label:       item.Label,
		Value:       item.Value,
		Color:       item.Color,
		Sort:        item.Sort,
		Status:      item.Status,
		Remark:      item.Remark,
		CreatedTime: formatTime(item.CreatedTime),
		UpdatedTime: formatTimePtr(item.UpdatedTime),
	}
}

func DictDataListFromModel(items []model.DictData) []DictDataDetail {
	result := make([]DictDataDetail, 0, len(items))
	for _, item := range items {
		result = append(result, DictDataFromModel(item))
	}
	return result
}
