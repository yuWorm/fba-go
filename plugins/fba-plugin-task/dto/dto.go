package dto

import (
	"encoding/json"
	"time"

	coretask "github.com/yuWorm/fba-go/core/task"
	"github.com/yuWorm/fba-plugin-task/model"
)

const TimeLayout = "2006-01-02 15:04:05"

type RegisteredTask struct {
	Name string `json:"name"`
	Task string `json:"task"`
}

type SchedulerParam struct {
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
}

type SchedulerDetail struct {
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

type TaskResultDetail struct {
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

type DeleteParam struct {
	PKs []int `json:"pks"`
}

func SchedulerFromModel(item model.TaskScheduler) SchedulerDetail {
	return SchedulerDetail{
		ID:             item.ID,
		Name:           item.Name,
		Task:           item.Task,
		Args:           decodeJSON(item.Args),
		Kwargs:         decodeJSON(item.Kwargs),
		Queue:          item.Queue,
		Exchange:       item.Exchange,
		RoutingKey:     item.RoutingKey,
		StartTime:      formatTimePtr(item.StartTime),
		ExpireTime:     formatTimePtr(item.ExpireTime),
		ExpireSeconds:  item.ExpireSeconds,
		Type:           item.Type,
		IntervalEvery:  item.IntervalEvery,
		IntervalPeriod: item.IntervalPeriod,
		Crontab:        item.Crontab,
		OneOff:         item.OneOff,
		Remark:         item.Remark,
		Enabled:        item.Enabled,
		TotalRunCount:  item.TotalRunCount,
		LastRunTime:    formatTimePtr(item.LastRunTime),
		CreatedTime:    formatTime(item.CreatedTime),
		UpdatedTime:    formatTimePtr(item.UpdatedTime),
	}
}

func SchedulersFromModel(items []model.TaskScheduler) []SchedulerDetail {
	result := make([]SchedulerDetail, 0, len(items))
	for _, item := range items {
		result = append(result, SchedulerFromModel(item))
	}
	return result
}

func TaskResultFromModel(item model.TaskResult) TaskResultDetail {
	return TaskResultDetail{
		ID:        item.ID,
		TaskID:    item.TaskID,
		Status:    string(coretask.MapAsynqState(item.Status)),
		Result:    decodeJSON(item.Result),
		DateDone:  formatTimePtr(item.DateDone),
		Traceback: item.Traceback,
		Name:      item.Name,
		Args:      decodeBytes(item.Args),
		Kwargs:    decodeBytes(item.Kwargs),
		Worker:    item.Worker,
		Retries:   item.Retries,
		Queue:     item.Queue,
	}
}

func TaskResultsFromModel(items []model.TaskResult) []TaskResultDetail {
	result := make([]TaskResultDetail, 0, len(items))
	for _, item := range items {
		result = append(result, TaskResultFromModel(item))
	}
	return result
}

func EncodeJSON(value any) *string {
	if value == nil {
		return nil
	}
	content, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	encoded := string(content)
	return &encoded
}

func decodeJSON(value *string) any {
	if value == nil {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(*value), &decoded); err != nil {
		return *value
	}
	return decoded
}

func decodeBytes(value []byte) any {
	if value == nil {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return string(value)
	}
	return decoded
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
