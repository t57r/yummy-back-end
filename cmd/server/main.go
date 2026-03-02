package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"

	"yummy/cmd/server/handlers"
	"yummy/cmd/server/httplog"
	"yummy/internal/config"
	"yummy/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	pool, err := db.CreatePool(cfg)
	if err != nil {
		slog.Error("db pool init failed", "error", err)
		os.Exit(1)
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
			if code >= fiber.StatusInternalServerError {
				httplog.Error(c, "request failed", err, "status", code)
			} else {
				httplog.Warn(c, "request rejected", "status", code, "message", msg)
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
	app.Use("/img", static.New("./img"))

	h.SetupRoutes(app, []byte(cfg.JWTAccessSecret))

	addr := ":" + cfg.Port
	slog.Info("server starting", "addr", addr)

	listenErrCh := make(chan error, 1)
	go func() {
		listenErrCh <- app.Listen(addr)
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-listenErrCh:
		if err != nil {
			slog.Error("server stopped", "error", err)
			os.Exit(1)
		}
		slog.Info("server stopped")
	case <-sigCtx.Done():
		slog.Info("shutdown signal received")
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			if errors.Is(err, fiber.ErrGracefulTimeout) {
				slog.Warn("graceful shutdown timeout reached", "timeout", "10s")
			} else {
				slog.Error("graceful shutdown failed", "error", err)
				os.Exit(1)
			}
		}

		if err := <-listenErrCh; err != nil {
			slog.Error("server stopped after shutdown with error", "error", err)
			os.Exit(1)
		}
		slog.Info("server stopped gracefully")
	}

}
