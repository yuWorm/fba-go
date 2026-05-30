package dto

type AuthLoginParam struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UUID     string `json:"uuid"`
	Captcha  string `json:"captcha"`
}

type CaptchaDetail struct {
	IsEnabled     bool   `json:"is_enabled"`
	ExpireSeconds int    `json:"expire_seconds"`
	UUID          string `json:"uuid"`
	Image         string `json:"image"`
}

type SwaggerToken struct {
	AccessToken string     `json:"access_token"`
	TokenType   string     `json:"token_type"`
	User        UserDetail `json:"user"`
}

type AccessTokenBase struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpireTime string `json:"access_token_expire_time"`
	SessionUUID           string `json:"session_uuid"`
}

type LoginToken struct {
	AccessTokenBase
	PasswordExpireDaysRemaining *int       `json:"password_expire_days_remaining"`
	User                        UserDetail `json:"user"`
}
