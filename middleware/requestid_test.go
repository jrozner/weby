package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value(RequestIDKey).(string); !ok {
			t.Errorf("request id not set")
		}
	})

	mw := RequestID(handler)

	mw.ServeHTTP(w, r)
}
