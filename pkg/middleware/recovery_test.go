package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverHandler(t *testing.T) {
	handler := RecoveryHandler()
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic("Unexpected error!")
	})

	recorder := httptest.NewRecorder()
	recovery := handler(handlerFunc)
	request, err := http.NewRequest(http.MethodGet, "/path/asdf", nil)
	if err != nil {
		t.Errorf("Got error %#v, expect no error", err)
	}

	recovery.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Got code %#v, wanted code %#v", recorder.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(recorder.Body.String(), "Unexpected error!") {
		t.Errorf("Got response %#v, wanted substring %#v", recorder.Body.String(), "Unexpected error!")
	}
}
