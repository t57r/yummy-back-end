package auth

import (
	"context"
	"errors"
	"strings"

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
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.Email == "" || len(req.Password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
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
			return fiber.NewError(fiber.StatusConflict, "email already exists")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	tokens, err := h.Auth.IssueTokens(ctx, createUserRow.ID)
	if err != nil {
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
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	ctx := context.Background()

	user, err := h.Auth.Queries.GetUserByEmail(ctx, req.Email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	tokens, err := h.Auth.IssueTokens(ctx, user.ID)
	if err != nil {
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
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing refresh_token")
	}

	tokens, err := h.Auth.Refresh(context.Background(), req.RefreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	return c.JSON(fiber.Map{"tokens": tokens})
}

func (h *Handler) Me(c fiber.Ctx) error {
	userIDAny := c.Locals("userID")
	userID, _ := userIDAny.(int64)

	user, err := h.Auth.Queries.GetUserByID(context.Background(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}

	return c.JSON(fiber.Map{"user": userDTO{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: utils.FormatTimestamptzOrEmpty(user.CreatedAt),
	}})
}
