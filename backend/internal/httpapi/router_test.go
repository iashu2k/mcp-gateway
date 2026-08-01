package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIRoot(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "MCP Gateway API",
		})
	})

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}