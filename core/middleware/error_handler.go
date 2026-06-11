package middleware

import (
	stderrors "errors"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/response"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	return NewErrorHandler(config.Options{})(c, err)
}

func NewErrorHandler(opts config.Options) fiber.ErrorHandler {
	opts = opts.WithDefaults()
	return func(c fiber.Ctx, err error) error {
		mapped := mapError(err)
		msg := mapped.message
		if opts.Middleware.ErrorResponse.IncludeDetail && mapped.status >= fiber.StatusInternalServerError && err != nil {
			msg = err.Error()
		}

		return c.Status(mapped.status).JSON(response.Error(mapped.code, msg, requestIDFromCtx(c)))
	}
}

type mappedError struct {
	status  int
	code    int
	message string
}

func mapError(err error) mappedError {
	mapped := mappedError{
		status:  fiber.StatusInternalServerError,
		code:    fiber.StatusInternalServerError,
		message: "内部服务器错误",
	}
	var appErr *fbaerrors.AppError
	if stderrors.As(err, &appErr) {
		mapped.status = appErr.HTTPStatus()
		mapped.code = appErr.Code()
		mapped.message = appErr.PublicMessage()
		return mapped
	}
	var fiberErr *fiber.Error
	if stderrors.As(err, &fiberErr) {
		mapped.status = fiberErr.Code
		mapped.code = fiberErr.Code
		mapped.message = fiberErr.Message
	}
	return mapped
}
