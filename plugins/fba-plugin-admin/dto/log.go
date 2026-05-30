package dto

import "github.com/yuWorm/fba-plugin-admin/model"

type UploadURL struct {
	URL string `json:"url"`
}

type LoginLogDetail struct {
	ID          int     `json:"id"`
	UserUUID    string  `json:"user_uuid"`
	Username    string  `json:"username"`
	Status      int     `json:"status"`
	IP          string  `json:"ip"`
	Country     *string `json:"country"`
	Region      *string `json:"region"`
	City        *string `json:"city"`
	UserAgent   *string `json:"user_agent"`
	Browser     *string `json:"browser"`
	OS          *string `json:"os"`
	Device      *string `json:"device"`
	Msg         string  `json:"msg"`
	LoginTime   string  `json:"login_time"`
	CreatedTime string  `json:"created_time"`
}

type OperaLogDetail struct {
	ID          int            `json:"id"`
	TraceID     string         `json:"trace_id"`
	Username    *string        `json:"username"`
	Method      string         `json:"method"`
	Title       string         `json:"title"`
	Path        string         `json:"path"`
	IP          string         `json:"ip"`
	Country     *string        `json:"country"`
	Region      *string        `json:"region"`
	City        *string        `json:"city"`
	UserAgent   *string        `json:"user_agent"`
	Browser     *string        `json:"browser"`
	OS          *string        `json:"os"`
	Device      *string        `json:"device"`
	Args        map[string]any `json:"args"`
	Status      int            `json:"status"`
	Code        string         `json:"code"`
	Msg         *string        `json:"msg"`
	CostTime    float64        `json:"cost_time"`
	OperaTime   string         `json:"opera_time"`
	CreatedTime string         `json:"created_time"`
}

func LoginLogFromModel(item model.LoginLog) LoginLogDetail {
	return LoginLogDetail{
		ID:          item.ID,
		UserUUID:    item.UserUUID,
		Username:    item.Username,
		Status:      item.Status,
		IP:          item.IP,
		Country:     item.Country,
		Region:      item.Region,
		City:        item.City,
		UserAgent:   item.UserAgent,
		Browser:     item.Browser,
		OS:          item.OS,
		Device:      item.Device,
		Msg:         item.Msg,
		LoginTime:   formatTime(item.LoginTime),
		CreatedTime: formatTime(item.CreatedTime),
	}
}

func LoginLogsFromModel(items []model.LoginLog) []LoginLogDetail {
	result := make([]LoginLogDetail, 0, len(items))
	for _, item := range items {
		result = append(result, LoginLogFromModel(item))
	}
	return result
}

func OperaLogFromModel(item model.OperaLog) OperaLogDetail {
	return OperaLogDetail{
		ID:          item.ID,
		TraceID:     item.TraceID,
		Username:    item.Username,
		Method:      item.Method,
		Title:       item.Title,
		Path:        item.Path,
		IP:          item.IP,
		Country:     item.Country,
		Region:      item.Region,
		City:        item.City,
		UserAgent:   item.UserAgent,
		Browser:     item.Browser,
		OS:          item.OS,
		Device:      item.Device,
		Args:        cloneAnyMap(item.Args),
		Status:      item.Status,
		Code:        item.Code,
		Msg:         item.Msg,
		CostTime:    item.CostTime,
		OperaTime:   formatTime(item.OperaTime),
		CreatedTime: formatTime(item.CreatedTime),
	}
}

func OperaLogsFromModel(items []model.OperaLog) []OperaLogDetail {
	result := make([]OperaLogDetail, 0, len(items))
	for _, item := range items {
		result = append(result, OperaLogFromModel(item))
	}
	return result
}

func cloneAnyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
