package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseAccessToken(t *testing.T) {
	secret := []byte("test-secret")

	makeToken := func(method jwt.SigningMethod, claims jwt.MapClaims) string {
		tok, err := jwt.NewWithClaims(method, claims).SignedString(secret)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}
		return tok
	}

	tests := []struct {
		name    string
		token   string
		wantID  int64
		wantErr bool
	}{
		{
			name: "valid access token",
			token: makeToken(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": int64(123),
				"typ": "access",
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantID:  123,
			wantErr: false,
		},
		{
			name: "wrong token type",
			token: makeToken(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": int64(123),
				"typ": "refresh",
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantErr: true,
		},
		{
			name: "missing subject",
			token: makeToken(jwt.SigningMethodHS256, jwt.MapClaims{
				"typ": "access",
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantErr: true,
		},
		{
			name: "non integer subject",
			token: makeToken(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": 1.5,
				"typ": "access",
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantErr: true,
		},
		{
			name: "negative subject",
			token: makeToken(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": float64(-1),
				"typ": "access",
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantErr: true,
		},
		{
			name: "wrong signing method",
			token: makeToken(jwt.SigningMethodHS512, jwt.MapClaims{
				"sub": int64(123),
				"typ": "access",
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := ParseAccessToken(tt.token, secret)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (id=%d)", gotID)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotID != tt.wantID {
				t.Fatalf("unexpected user id: got=%d want=%d", gotID, tt.wantID)
			}
		})
	}
}
