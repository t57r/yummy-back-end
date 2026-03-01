package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"yummy/cmd/server/handlers"
	"yummy/internal/config"
	"yummy/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	pool, err := db.CreatePool(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	queries := db.New(pool)

	h := handlers.NewHandlers(cfg, queries)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal error"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				msg = e.Message
			}
			return c.Status(code).JSON(fiber.Map{"error": msg})
		},
	})

	// Enable CORS with custom configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:8082"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
	}))

	h.SetupRoutes(app, []byte(cfg.JWTAccessSecret))

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
