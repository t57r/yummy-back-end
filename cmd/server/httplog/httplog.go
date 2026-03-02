package httplog

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

func Error(c fiber.Ctx, msg string, err error, attrs ...any) {
	args := baseAttrs(c)
	if err != nil {
		args = append(args, "error", err)
	}
	args = append(args, attrs...)
	slog.Error(msg, args...)
}

func Warn(c fiber.Ctx, msg string, attrs ...any) {
	args := baseAttrs(c)
	args = append(args, attrs...)
	slog.Warn(msg, args...)
}

func baseAttrs(c fiber.Ctx) []any {
	return []any{
		"request_id", c.RequestID(),
		"method", c.Method(),
		"path", c.Path(),
	}
}
