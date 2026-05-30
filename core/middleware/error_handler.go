package middleware

import (
	stderrors "errors"

	"github.com/gofiber/fiber/v3"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/response"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	code := fiber.StatusInternalServerError
	msg := "内部服务器错误"

	var appErr *fbaerrors.AppError
	if stderrors.As(err, &appErr) {
		status = appErr.HTTPStatus()
		code = appErr.Code()
		msg = appErr.PublicMessage()
	} else {
		var fiberErr *fiber.Error
		if stderrors.As(err, &fiberErr) {
			status = fiberErr.Code
			code = fiberErr.Code
			msg = fiberErr.Message
		}
	}

	return c.Status(status).JSON(response.Error(code, msg, requestIDFromCtx(c)))
}
