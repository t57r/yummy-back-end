package categories

import (
	"context"

	"yummy/internal/db"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	Queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{Queries: queries}
}

func (h *Handler) List(c fiber.Ctx) error {
	categories, err := h.Queries.ListCategories(context.Background())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	return c.JSON(fiber.Map{"data": categories})
}
