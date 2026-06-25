package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/response"
	"offer-catcher/server/internal/service"
	"offer-catcher/server/internal/session"
)

type contextKey string

const (
	userIDKey    contextKey = "user_id"
	sessionIDKey contextKey = "session_id"
)

// Auth returns middleware that validates JWT token and Redis session.
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
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

			tokenString := parts[1]
			userID, sessionID, err := authService.ParseToken(tokenString)
			if err != nil {
				log.Printf("[auth] token parse error: %v", err)
				response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
				return
			}

			// Validate Redis session with current-session check
			if err := authService.ValidateSession(r.Context(), userID, sessionID); err != nil {
				if errors.Is(err, session.ErrSessionReplaced) {
					log.Printf("[auth] session replaced: user=%d", userID)
					response.Error(w, apperror.ErrSessionReplaced.HTTPStatus, apperror.ErrSessionReplaced.Code, apperror.ErrSessionReplaced.Message)
					return
				}
				log.Printf("[auth] session invalid: %v", err)
				response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, sessionIDKey, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDContextKey is the exported context key for user ID — used by router dispatch.
var UserIDContextKey = userIDKey
func UserIDFromContext(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(userIDKey).(uint64)
	return id, ok
}

// SessionIDFromContext extracts the session ID from the request context.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(sessionIDKey).(string)
	return id, ok
}
