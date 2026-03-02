package recipes

import (
	"context"
	"strconv"
	"strings"

	"yummy/cmd/server/httplog"
	"yummy/internal/db"
	"yummy/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	Queries *db.Queries
}

type pagination struct {
	Limit      int   `json:"limit"`
	LastID     int64 `json:"last_id"`
	NextLastID int64 `json:"next_last_id,omitempty"`
	HasNext    bool  `json:"has_next"`
	Total      int64 `json:"total"`
}

type listResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination pagination `json:"pagination"`
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{Queries: queries}
}

func (h *Handler) List(c fiber.Ctx) error {
	ctx := context.Background()

	category := strings.TrimSpace(c.Query("category"))

	limit := clampInt(parseInt(c.Query("limit"), 20), 1, 100)
	lastID := parseInt64(c.Query("last_id"), 0)
	if lastID < 0 {
		httplog.Warn(c, "recipes list invalid last_id", "last_id_raw", c.Query("last_id"))
		return fiber.NewError(fiber.StatusBadRequest, "invalid last_id")
	}
	queryLimit := limit + 1

	qText := utils.ToPgText(c.Query("q"))
	lastIDArg := pgtype.Int8{Int64: lastID, Valid: lastID > 0}

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
			httplog.Error(c, "recipes count by category failed", err, "category", category)
			return fiber.NewError(fiber.StatusInternalServerError, "db error")
		}

		rows, err = h.Queries.ListRecipesByCategoryName(ctx, db.ListRecipesByCategoryNameParams{
			Name:   category,
			Limit:  int32(queryLimit),
			LastID: lastIDArg,
			Q:      qText,
		})
		if err != nil {
			httplog.Error(c, "recipes list by category failed", err, "category", category, "last_id", lastID, "limit", limit)
			return fiber.NewError(fiber.StatusInternalServerError, "db error")
		}
	} else {
		total, err = h.Queries.CountRecipes(ctx, qText)
		if err != nil {
			httplog.Error(c, "recipes count failed", err)
			return fiber.NewError(fiber.StatusInternalServerError, "db error")
		}

		rows, err = h.Queries.ListRecipes(ctx, db.ListRecipesParams{
			Limit:  int32(queryLimit),
			LastID: lastIDArg,
			Q:      qText,
		})
		if err != nil {
			httplog.Error(c, "recipes list failed", err, "last_id", lastID, "limit", limit)
			return fiber.NewError(fiber.StatusInternalServerError, "db error")
		}
	}

	hasNext := len(rows) > limit
	if hasNext {
		rows = rows[:limit]
	}

	var nextLastID int64
	if len(rows) > 0 {
		nextLastID = rows[len(rows)-1].ID
	}

	return c.JSON(listResponse[db.Recipe]{
		Data: rows,
		Pagination: pagination{
			Limit:      limit,
			LastID:     lastID,
			NextLastID: nextLastID,
			HasNext:    hasNext,
			Total:      total,
		},
	})
}

func (h *Handler) GetByID(c fiber.Ctx) error {
	ctx := context.Background()

	recipeID, err := strconv.ParseInt(c.Params("recipeId"), 10, 64)
	if err != nil || recipeID <= 0 {
		httplog.Warn(c, "recipes get invalid recipeId", "recipe_id_raw", c.Params("recipeId"))
		return fiber.NewError(fiber.StatusBadRequest, "invalid recipeId")
	}

	recipe, err := h.Queries.GetRecipeByID(ctx, recipeID)
	if err != nil {
		httplog.Warn(c, "recipes get not found", "recipe_id", recipeID)
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

func parseInt64(s string, def int64) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
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
