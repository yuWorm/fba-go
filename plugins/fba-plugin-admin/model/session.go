package model

import "time"

type Session struct {
	ID            int
	SessionUUID   string
	Username      string
	Nickname      string
	IP            string
	OS            string
	Browser       string
	Device        string
	Status        int
	LastLoginTime string
	ExpireTime    time.Time
}
