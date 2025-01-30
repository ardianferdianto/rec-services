package middleware

import (
	"fmt"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"github.com/ardianferdianto/reconciliation-service/pkg/response"
	"log/slog"
	"net/http"
)

func RecoveryHandler() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := recover()
				if err != nil {
					slog.ErrorContext(r.Context(), "Panic recovered", logger.ErrAttr(fmt.Errorf("recovery: %v", err)))
					jsonBody := response.BuildErrorResponse(fmt.Sprintf("%v", err))
					response.WriteJSON(r.Context(), w, http.StatusInternalServerError, jsonBody)
				}
			}()

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
