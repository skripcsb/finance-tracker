package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey int

const keyUserID contextKey = 0

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    12 * time.Hour,
	}
}

func (tm *TokenManager) Issue(userID int) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"uid": userID,
		"exp": now.Add(tm.ttl).Unix(),
		"iat": now.Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tm.secret)
}

func (tm *TokenManager) Verify(raw string) (int, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return tm.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("bad token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("bad claims")
	}

	uid, ok := claims["uid"].(float64)
	if !ok {
		return 0, fmt.Errorf("no uid in token")
	}
	return int(uid), nil
}

func (tm *TokenManager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		uid, err := tm.Verify(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), keyUserID, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) int {
	uid, _ := ctx.Value(keyUserID).(int)
	return uid
}
