package response

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ardianferdianto/reconciliation-service/pkg/logger"
	"log/slog"
	"net/http"
)

type GenericResponse struct {
	Success bool    `json:"success"`
	Error   *string `json:"error,omitempty"`
}

func WriteJSONString(w http.ResponseWriter, statusCode int, v string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(v))
}

func WriteJSON(ctx context.Context, w http.ResponseWriter, statusCode int, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := enc.Encode(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(ctx, "failed to marshal json response", logger.ErrAttr(err))
	}
	w.Write(buf.Bytes())
}

func BuildErrorResponse(message string) *GenericResponse {
	return &GenericResponse{
		Success: false,
		Error:   &message,
	}
}

func GenericSuccessResponse() *GenericResponse {
	return &GenericResponse{
		Success: true,
	}
}
