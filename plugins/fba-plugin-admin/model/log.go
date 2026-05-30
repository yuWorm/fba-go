package model

import "time"

type LoginLog struct {
	ID          int
	UserUUID    string
	Username    string
	Status      int
	IP          string
	Country     *string
	Region      *string
	City        *string
	UserAgent   *string
	Browser     *string
	OS          *string
	Device      *string
	Msg         string
	LoginTime   time.Time
	CreatedTime time.Time
}

type OperaLog struct {
	ID          int
	TraceID     string
	Username    *string
	Method      string
	Title       string
	Path        string
	IP          string
	Country     *string
	Region      *string
	City        *string
	UserAgent   *string
	Browser     *string
	OS          *string
	Device      *string
	Args        map[string]any
	Status      int
	Code        string
	Msg         *string
	CostTime    float64
	OperaTime   time.Time
	CreatedTime time.Time
}
