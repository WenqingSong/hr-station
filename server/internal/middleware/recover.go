package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/response"
)

// Recover catches panics in downstream handlers, logs the stack trace,
// and returns a unified error response without exposing panic details to the client.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v\n%s", err, debug.Stack())
				response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
