package response

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSONString(t *testing.T) {
	tr := httptest.NewRecorder()
	WriteJSONString(tr, http.StatusInternalServerError, `{"success":true}`)

	assert.Equal(t, `{"success":true}`, strings.TrimSpace(tr.Body.String()))
	assert.Equal(t, http.StatusInternalServerError, tr.Code)
	assert.Equal(t, "application/json; charset=utf-8", tr.Header().Get("Content-Type"))
}

func TestWriteJSON(t *testing.T) {
	tr := httptest.NewRecorder()
	WriteJSON(context.Background(), tr, http.StatusInternalServerError, BuildErrorResponse("message"))

	assert.Equal(t, `{"success":false,"error":"message"}`, strings.TrimSpace(tr.Body.String()))
	assert.Equal(t, http.StatusInternalServerError, tr.Code)
	assert.Equal(t, "application/json; charset=utf-8", tr.Header().Get("Content-Type"))
}

func TestGenericSuccessResponse(t *testing.T) {
	tr := httptest.NewRecorder()
	WriteJSON(context.Background(), tr, http.StatusOK, GenericSuccessResponse())

	assert.Equal(t, `{"success":true}`, strings.TrimSpace(tr.Body.String()))
	assert.Equal(t, http.StatusOK, tr.Code)
	assert.Equal(t, "application/json; charset=utf-8", tr.Header().Get("Content-Type"))
}
