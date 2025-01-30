package rest

import (
	"github.com/ardianferdianto/reconciliation-service/pkg/response"
	"net/http"
)

func NotFoundHandler(w http.ResponseWriter, req *http.Request) {
	response.WriteJSON(req.Context(), w, http.StatusNotFound,
		response.BuildErrorResponse("API not found"))
}

func PingHandler(w http.ResponseWriter, req *http.Request) {
	response.WriteJSON(req.Context(), w, http.StatusOK, response.GenericSuccessResponse())
}
