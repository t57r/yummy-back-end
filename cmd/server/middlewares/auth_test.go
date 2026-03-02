package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"yummy/cmd/server/userctx"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func TestRequireAuth(t *testing.T) {
	secret := []byte("test-secret")

	makeAccessToken := func(userID int64) string {
		tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userID,
			"typ": "access",
			"exp": time.Now().Add(time.Hour).Unix(),
		}).SignedString(secret)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}
		return tok
	}

	app := fiber.New()
	app.Use(RequireAuth(secret))
	app.Get("/protected", func(c fiber.Ctx) error {
		id, _ := c.Locals(userctx.UserIDLocalKey).(int64)
		return c.JSON(fiber.Map{"id": id})
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("unexpected status: got=%d want=%d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Token abc")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("unexpected status: got=%d want=%d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("valid bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+makeAccessToken(42))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: got=%d want=%d", resp.StatusCode, http.StatusOK)
		}

		var body struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if body.ID != 42 {
			t.Fatalf("unexpected user id in locals: got=%d want=%d", body.ID, 42)
		}
	})
}
