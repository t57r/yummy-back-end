package middlewares

import (
	"strings"
	"yummy/cmd/server/handlers/auth"

	"github.com/gofiber/fiber/v3"
)

func RequireAuth(accessSecret []byte) fiber.Handler {
	return func(c fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing Authorization header")
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid Authorization header")
		}
		userID, err := auth.ParseAccessToken(parts[1], accessSecret)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}
		c.Locals("userID", userID)
		return c.Next()
	}
}
