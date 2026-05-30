package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
)

func Recover() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = fbaerrors.New(
					fiber.StatusInternalServerError,
					fiber.StatusInternalServerError,
					"内部服务器错误",
					fmt.Errorf("panic: %v", recovered),
				)
			}
		}()
		return c.Next()
	}
}
