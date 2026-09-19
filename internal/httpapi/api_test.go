package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthIncludesRequestID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	requestID(http.HandlerFunc((&API{}).health)).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("request ID missing")
	}
}
