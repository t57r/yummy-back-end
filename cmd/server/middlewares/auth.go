package middlewares

import (
	"strings"
	"yummy/cmd/server/handlers/auth"
	"yummy/cmd/server/httplog"
	"yummy/cmd/server/userctx"

	"github.com/gofiber/fiber/v3"
)

func RequireAuth(accessSecret []byte) fiber.Handler {
	return func(c fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			httplog.Warn(c, "auth missing Authorization header")
			return fiber.NewError(fiber.StatusUnauthorized, "missing Authorization header")
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httplog.Warn(c, "auth invalid Authorization header")
			return fiber.NewError(fiber.StatusUnauthorized, "invalid Authorization header")
		}
		userID, err := auth.ParseAccessToken(parts[1], accessSecret)
		if err != nil {
			httplog.Warn(c, "auth invalid access token")
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}
		c.Locals(userctx.UserIDLocalKey, userID)
		return c.Next()
	}
}
