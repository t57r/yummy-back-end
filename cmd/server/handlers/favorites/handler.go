package favorites

import (
	"context"
	"strconv"

	"yummy/cmd/server/httplog"
	"yummy/cmd/server/userctx"
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
	userID, err := userctx.CurrentUserID(c)
	if err != nil {
		httplog.Warn(c, "favorites add unauthorized context")
		return err
	}
	recipeID, err := strconv.ParseInt(c.Params("recipeId"), 10, 64)
	if err != nil {
		httplog.Warn(c, "favorites add invalid recipeId", "recipe_id_raw", c.Params("recipeId"))
		return fiber.NewError(fiber.StatusBadRequest, "invalid recipeId")
	}

	// Ensure recipe exists
	exists, err := h.Queries.RecipeExists(context.Background(), recipeID)
	if err != nil {
		httplog.Error(c, "favorites add recipe exists check failed", err, "recipe_id", recipeID, "user_id", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	if !exists {
		httplog.Warn(c, "favorites add recipe not found", "recipe_id", recipeID, "user_id", userID)
		return fiber.NewError(fiber.StatusNotFound, "recipe not found")
	}

	err = h.Queries.AddFavorite(context.Background(), db.AddFavoriteParams{
		UserID:   userID,
		RecipeID: recipeID,
	})
	if err != nil {
		httplog.Error(c, "favorites add failed", err, "recipe_id", recipeID, "user_id", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Remove(c fiber.Ctx) error {
	userID, err := userctx.CurrentUserID(c)
	if err != nil {
		httplog.Warn(c, "favorites remove unauthorized context")
		return err
	}
	recipeID, err := strconv.ParseInt(c.Params("recipeId"), 10, 64)
	if err != nil {
		httplog.Warn(c, "favorites remove invalid recipeId", "recipe_id_raw", c.Params("recipeId"))
		return fiber.NewError(fiber.StatusBadRequest, "invalid recipeId")
	}

	err = h.Queries.RemoveFavorite(context.Background(), db.RemoveFavoriteParams{
		UserID:   userID,
		RecipeID: recipeID,
	})
	if err != nil {
		httplog.Error(c, "favorites remove failed", err, "recipe_id", recipeID, "user_id", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) List(c fiber.Ctx) error {
	userID, err := userctx.CurrentUserID(c)
	if err != nil {
		httplog.Warn(c, "favorites list unauthorized context")
		return err
	}

	var limit int32 = 20
	var offset int32 = 0

	total, err := h.Queries.CountFavoritesByUser(context.Background(), userID)
	if err != nil {
		httplog.Error(c, "favorites count failed", err, "user_id", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	recipes, err := h.Queries.ListFavoriteRecipesByUser(context.Background(), db.ListFavoriteRecipesByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		httplog.Error(c, "favorites list failed", err, "user_id", userID)
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
