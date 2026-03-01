package favorites

import (
	"context"
	"strconv"

	"yummy/internal/db"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	Queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{Queries: queries}
}

func (h *Handler) Add(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	recipeID, err := strconv.ParseInt(c.Params("recipeId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid recipeId")
	}

	// Ensure recipe exists
	exists, err := h.Queries.RecipeExists(context.Background(), recipeID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "recipe not found")
	}

	err = h.Queries.AddFavorite(context.Background(), db.AddFavoriteParams{
		UserID:   userID,
		RecipeID: recipeID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Remove(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	recipeID, err := strconv.ParseInt(c.Params("recipeId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid recipeId")
	}

	err = h.Queries.RemoveFavorite(context.Background(), db.RemoveFavoriteParams{
		UserID:   userID,
		RecipeID: recipeID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) List(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	var limit int32 = 20
	var offset int32 = 0

	total, err := h.Queries.CountFavoritesByUser(context.Background(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	recipes, err := h.Queries.ListFavoriteRecipesByUser(context.Background(), db.ListFavoriteRecipesByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	return c.JSON(fiber.Map{
		"data": recipes,
		"pagination": fiber.Map{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}
