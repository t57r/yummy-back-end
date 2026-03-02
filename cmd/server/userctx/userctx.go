package userctx

import "github.com/gofiber/fiber/v3"

const UserIDLocalKey = "userID"

func CurrentUserID(c fiber.Ctx) (int64, error) {
	userIDAny := c.Locals(UserIDLocalKey)
	userID, ok := userIDAny.(int64)
	if !ok || userID <= 0 {
		return 0, fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	return userID, nil
}
