package recipes

import (
	"context"
	"strconv"
	"strings"

	"yummy/internal/db"
	"yummy/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Queries *db.Queries
}

type pagination struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}

type listResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination pagination `json:"pagination"`
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{Queries: queries}
}

func (h *Handler) List(c *fiber.Ctx) error {
	ctx := context.Background()

	category := strings.TrimSpace(c.Query("category"))

	limit := clampInt(parseInt(c.Query("limit"), 20), 1, 100)
	offset := clampInt(parseInt(c.Query("offset"), 0), 0, 1_000_000)

	qText := utils.ToPgText(c.Query("q"))

	var (
		total int64
		rows  []db.Recipe
		err   error
	)

	if category != "" {
		total, err = h.Queries.CountRecipesByCategoryName(ctx, db.CountRecipesByCategoryNameParams{
			Name: category,
			Q:    qText,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "db error")
		}

		rows, err = h.Queries.ListRecipesByCategoryName(ctx, db.ListRecipesByCategoryNameParams{
			Name:   category,
			Limit:  int32(limit),
			Offset: int32(offset),
			Q:      qText,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "db error")
		}
	} else {
		total, err = h.Queries.CountRecipes(ctx, qText)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "db error")
		}

		rows, err = h.Queries.ListRecipes(ctx, db.ListRecipesParams{
			Limit:  int32(limit),
			Offset: int32(offset),
			Q:      qText,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "db error")
		}
	}

	return c.JSON(listResponse[db.Recipe]{
		Data: rows,
		Pagination: pagination{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	ctx := context.Background()

	recipeID, err := strconv.ParseInt(c.Params("recipeId"), 10, 64)
	if err != nil || recipeID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid recipeId")
	}

	recipe, err := h.Queries.GetRecipeByID(ctx, recipeID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}

	return c.JSON(fiber.Map{"data": recipe})
}

func parseInt(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
