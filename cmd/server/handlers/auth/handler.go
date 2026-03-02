package auth

import (
	"context"
	"errors"
	"strings"

	"yummy/cmd/server/httplog"
	"yummy/cmd/server/userctx"
	"yummy/internal/db"
	"yummy/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Auth *Service
}

type signUpReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type userDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{Auth: service}
}

func (h *Handler) SignUp(c fiber.Ctx) error {
	var req signUpReq
	if err := c.Bind().Body(&req); err != nil {
		httplog.Warn(c, "signup invalid json", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.Email == "" || len(req.Password) < 8 {
		httplog.Warn(c, "signup invalid input", "email", req.Email)
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httplog.Error(c, "password hash failed", err)
		return fiber.NewError(fiber.StatusInternalServerError, "hash failed")
	}
	ctx := context.Background()

	createUserRow, err := h.Auth.Queries.CreateUser(ctx, db.CreateUserParams{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashBytes),
	})
	if err != nil {
		// unique email constraint
		if strings.Contains(err.Error(), "duplicate key") {
			httplog.Warn(c, "signup email already exists", "email", req.Email)
			return fiber.NewError(fiber.StatusConflict, "email already exists")
		}
		httplog.Error(c, "signup db create user failed", err, "email", req.Email)
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	tokens, err := h.Auth.IssueTokens(ctx, createUserRow.ID)
	if err != nil {
		httplog.Error(c, "signup token issue failed", err, "user_id", createUserRow.ID)
		return fiber.NewError(fiber.StatusInternalServerError, "token issue failed")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":   userDTO{ID: createUserRow.ID, Name: req.Name, Email: req.Email, CreatedAt: utils.FormatTimestamptzOrEmpty(createUserRow.CreatedAt)},
		"tokens": tokens,
	})
}

func (h *Handler) SignIn(c fiber.Ctx) error {
	var req signInReq
	if err := c.Bind().Body(&req); err != nil {
		httplog.Warn(c, "signin invalid json", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	ctx := context.Background()

	user, err := h.Auth.Queries.GetUserByEmail(ctx, req.Email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httplog.Warn(c, "signin unknown email", "email", req.Email)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}
		httplog.Error(c, "signin db lookup failed", err, "email", req.Email)
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		httplog.Warn(c, "signin invalid password", "email", req.Email)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	tokens, err := h.Auth.IssueTokens(ctx, user.ID)
	if err != nil {
		httplog.Error(c, "signin token issue failed", err, "user_id", user.ID)
		return fiber.NewError(fiber.StatusInternalServerError, "token issue failed")
	}

	return c.JSON(fiber.Map{
		"user": userDTO{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: utils.FormatTimestamptzOrEmpty(user.CreatedAt),
		},
		"tokens": tokens,
	})

}

func (h *Handler) Refresh(c fiber.Ctx) error {
	var req refreshReq
	if err := c.Bind().Body(&req); err != nil {
		httplog.Warn(c, "refresh invalid json", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		httplog.Warn(c, "refresh token missing")
		return fiber.NewError(fiber.StatusBadRequest, "missing refresh_token")
	}

	tokens, err := h.Auth.Refresh(context.Background(), req.RefreshToken)
	if err != nil {
		httplog.Warn(c, "refresh token rejected")
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	return c.JSON(fiber.Map{"tokens": tokens})
}

func (h *Handler) Me(c fiber.Ctx) error {
	userID, err := userctx.CurrentUserID(c)
	if err != nil {
		httplog.Warn(c, "me unauthorized context")
		return err
	}

	user, err := h.Auth.Queries.GetUserByID(context.Background(), userID)
	if err != nil {
		httplog.Warn(c, "me user not found", "user_id", userID)
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}

	return c.JSON(fiber.Map{"user": userDTO{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: utils.FormatTimestamptzOrEmpty(user.CreatedAt),
	}})
}
