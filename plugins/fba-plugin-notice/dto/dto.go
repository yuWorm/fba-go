package dto

import (
	"time"

	"github.com/yuWorm/fba-plugin-notice/model"
)

const TimeLayout = "2006-01-02 15:04:05"

type NoticeParam struct {
	Title   string `json:"title"`
	Type    int    `json:"type"`
	Status  int    `json:"status"`
	Content string `json:"content"`
}

type DeleteParam struct {
	PKs []int `json:"pks"`
}

type NoticeDetail struct {
	Title       string  `json:"title"`
	Type        int     `json:"type"`
	Status      int     `json:"status"`
	Content     string  `json:"content"`
	ID          int     `json:"id"`
	CreatedTime string  `json:"created_time"`
	UpdatedTime *string `json:"updated_time"`
}

func NoticeFromModel(item model.Notice) NoticeDetail {
	return NoticeDetail{
		ID:          item.ID,
		Title:       item.Title,
		Type:        item.Type,
		Status:      item.Status,
		Content:     item.Content,
		CreatedTime: formatTime(item.CreatedTime),
		UpdatedTime: formatTimePtr(item.UpdatedTime),
	}
}

func NoticesFromModel(items []model.Notice) []NoticeDetail {
	result := make([]NoticeDetail, 0, len(items))
	for _, item := range items {
		result = append(result, NoticeFromModel(item))
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
