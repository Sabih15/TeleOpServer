package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sabih15/TeleOpServer/internal/auth"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
)

type contextKey string

const userIDKey contextKey = "user_id"

// Auth returns a Chi middleware that validates the Bearer JWT and injects the user ID into context.
func Auth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			claims, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "), cfg.JWT.Secret)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user ID from context.
func UserIDFromContext(ctx context.Context) uint {
	if id, ok := ctx.Value(userIDKey).(uint); ok {
		return id
	}
	return 0
}
