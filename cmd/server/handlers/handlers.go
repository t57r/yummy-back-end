package handlers

import (
	"time"

	"yummy/cmd/server/handlers/auth"
	"yummy/cmd/server/handlers/categories"
	"yummy/cmd/server/handlers/favorites"
	"yummy/cmd/server/handlers/recipes"
	"yummy/cmd/server/middlewares"
	"yummy/internal/config"
	"yummy/internal/db"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	Auth       *auth.Handler
	Categories *categories.Handler
	Favorites  *favorites.Handler
	Recipes    *recipes.Handler
}

func NewHandlers(config *config.Config, queries *db.Queries) *Handlers {
	authService := auth.NewService(
		queries,
		config.JWTAccessSecret,
		config.JWTRefreshSecret,
		time.Duration(config.AccessTTLSeconds)*time.Second,
		time.Duration(config.RefreshTTLSeconds)*time.Second,
	)

	return &Handlers{
		Auth:       auth.NewHandler(authService),
		Categories: categories.NewHandler(queries),
		Favorites:  favorites.NewHandler(queries),
		Recipes:    recipes.NewHandler(queries),
	}
}

func (h *Handlers) SetupRoutes(app *fiber.App, accessSecret []byte) {
	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })

	app.Post("/auth/signup", h.Auth.SignUp)
	app.Post("/auth/signin", h.Auth.SignIn)
	app.Post("/auth/refresh", h.Auth.Refresh)

	app.Get("/categories", h.Categories.List)

	app.Get("/recipes", h.Recipes.List)
	app.Get("/recipes/:recipeId", h.Recipes.GetByID)

	authed := app.Group("", middlewares.RequireAuth(accessSecret))
	authed.Get("/me", h.Auth.Me)

	authed.Get("/me/favorites", h.Favorites.List)
	authed.Post("/me/favorites/:recipeId", h.Favorites.Add)
	authed.Delete("/me/favorites/:recipeId", h.Favorites.Remove)
}
