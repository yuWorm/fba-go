package api

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/response"
	"github.com/yuWorm/fba-plugin-email/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) Handler {
	if svc == nil {
		svc = service.New(service.Options{})
	}
	return Handler{service: svc}
}

func (h Handler) SendCaptcha(c fiber.Ctx) error {
	var param sendCaptchaParam
	if err := c.Bind().Body(&param); err != nil {
		return err
	}
	if err := h.service.SendCaptcha(c.RequestCtx(), []string(param.Recipients), c.IP()); err != nil {
		return err
	}
	return c.JSON(response.Success[any](nil))
}

type sendCaptchaParam struct {
	Recipients recipients `json:"recipients"`
}

type recipients []string

func (r *recipients) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*r = []string{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return fmt.Errorf("recipients must be a string or string array: %w", err)
	}
	*r = many
	return nil
}
