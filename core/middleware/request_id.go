package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v3"
)

const (
	RequestIDHeader   = "X-Request-ID"
	RequestIDLocalKey = "request_id"
)

func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Locals(RequestIDLocalKey, requestID)
		c.Set(RequestIDHeader, requestID)
		return c.Next()
	}
}

func RequestIDFromCtx(c fiber.Ctx) string {
	return requestIDFromCtx(c)
}

func requestIDFromCtx(c fiber.Ctx) string {
	if value, ok := c.Locals(RequestIDLocalKey).(string); ok && value != "" {
		return value
	}
	return newRequestID()
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "trace-unavailable"
	}
	return hex.EncodeToString(b[:])
}
