package middleware

import (
	"context"
	"net/http"
	"strings"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/response"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const userIDContextKey contextKey = "user_id"

// TokenParser is an interface for parsing JWT tokens.
type TokenParser interface {
	ParseToken(tokenString string) (uint64, error)
}

// Auth returns middleware that validates JWT Bearer tokens
// and puts the user_id into the request context.
func Auth(tokenParser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
				return
			}

			userID, err := tokenParser.ParseToken(parts[1])
			if err != nil {
				response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the user ID from the request context.
func UserIDFromContext(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uint64)
	return userID, ok
}
