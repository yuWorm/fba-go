package model

import "time"

type User struct {
	ID            int
	UUID          string
	DeptID        *int
	Username      string
	Nickname      string
	Password      string
	Avatar        *string
	Email         *string
	Phone         *string
	Status        int
	IsSuperuser   bool
	IsStaff       bool
	IsMultiLogin  bool
	JoinTime      time.Time
	LastLoginTime *time.Time
	DeletedTime   *time.Time
}
