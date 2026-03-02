package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"time"

	"yummy/internal/db"
	"yummy/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	Queries       *db.Queries
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func NewService(queries *db.Queries, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		Queries:       queries,
		AccessSecret:  []byte(accessSecret),
		RefreshSecret: []byte(refreshSecret),
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	}
}

func (s *Service) IssueTokens(ctx context.Context, userID int64) (Tokens, error) {
	accessExp := time.Now().Add(s.AccessTTL)
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": accessExp.Unix(),
		"typ": "access",
	}).SignedString(s.AccessSecret)
	if err != nil {
		return Tokens{}, err
	}

	// Refresh token: generate random, store hash in DB (revokable)
	refreshRaw, err := randomToken(48)
	if err != nil {
		return Tokens{}, err
	}
	refreshHash := hashToken(refreshRaw)
	refreshExp := time.Now().Add(s.RefreshTTL)

	err = s.Queries.InsertRefreshToken(ctx, db.InsertRefreshTokenParams{
		UserID:    userID,
		TokenHash: refreshHash,
		ExpiresAt: utils.TimeToPgTimestamptz(refreshExp),
	})
	if err != nil {
		return Tokens{}, err
	}

	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.AccessTTL.Seconds()),
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	h := hashToken(refreshToken)

	token, err := s.Queries.GetRefreshTokenByHash(ctx, h)
	if err != nil {
		return Tokens{}, errors.New("invalid refresh token")
	}

	// check revoked
	if token.RevokedAt.Valid {
		return Tokens{}, errors.New("refresh token revoked")
	}

	// check expiration
	if !token.ExpiresAt.Valid {
		return Tokens{}, errors.New("invalid refresh token expiration")
	}
	if time.Now().After(token.ExpiresAt.Time) {
		return Tokens{}, errors.New("refresh token expired")
	}

	// rotate (revoke old one)
	if err := s.Queries.RevokeRefreshTokenByHash(ctx, h); err != nil {
		return Tokens{}, err
	}

	return s.IssueTokens(ctx, token.UserID)
}

func ParseAccessToken(tokenStr string, accessSecret []byte) (int64, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return accessSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !tok.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	typ, _ := claims["typ"].(string)
	if typ != "access" {
		return 0, errors.New("wrong token type")
	}

	subFloat, ok := claims["sub"].(float64) // JSON numbers decode as float64
	if !ok {
		return 0, errors.New("missing sub")
	}
	if subFloat <= 0 || math.Trunc(subFloat) != subFloat {
		return 0, errors.New("invalid sub")
	}

	return int64(subFloat), nil
}

func randomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
