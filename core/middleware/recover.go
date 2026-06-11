package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
)

const (
	PanicLocalKey      = "panic"
	PanicStackLocalKey = "panic_stack"
)

type RecoverConfig struct {
	EnableStackTrace bool
}

func Recover(config ...RecoverConfig) fiber.Handler {
	cfg := RecoverConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				c.Locals(PanicLocalKey, fmt.Sprint(recovered))
				if cfg.EnableStackTrace {
					// Store the stack on the request context so the error logger can
					// correlate it with method, path, status, and trace_id.
					c.Locals(PanicStackLocalKey, string(debug.Stack()))
				}
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
