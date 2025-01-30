package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/ardianferdianto/reconciliation-service/pkg/contextprop"
	"github.com/ardianferdianto/reconciliation-service/pkg/response"
	"net/http"
	"strings"
)

const (
	AuthorizationHeader = "Authorization"
)

func respAuthFailed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	errorResponse := response.BuildErrorResponse("Authorization Failed")
	errorResponseBytes, _ := json.Marshal(errorResponse)
	w.Write(errorResponseBytes)
}

func BasicAuthMiddleware(authorizedClient map[string]string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := strings.SplitN(r.Header.Get(AuthorizationHeader), " ", 2)

			if len(auth) != 2 || auth[0] != "Basic" {
				respAuthFailed(w)
				return
			}

			payload, _ := base64.StdEncoding.DecodeString(auth[1])
			pair := strings.SplitN(string(payload), ":", 2)

			if len(pair) != 2 {
				respAuthFailed(w)
				return
			}

			clientID := pair[0]
			clientSecret := pair[1]

			if clientID == "" || clientSecret == "" {
				respAuthFailed(w)
				return
			}

			if authorizedClient[clientID] != clientSecret {
				respAuthFailed(w)
				return
			}

			ctx := context.WithValue(r.Context(), contextprop.ClientIDKey, clientID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
