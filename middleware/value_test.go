package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jrozner/weby/contextvalue"
)

var (
	valueTestKey   contextvalue.Key = "hi"
	valueTestValue                  = "hi"
)

func TestValue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			v   string
			err error
		)

		if v, err = contextvalue.Value[string](r.Context(), valueTestKey); err != nil {
			t.Errorf("error getting value: %s", err)
		}

		if v != valueTestValue {
			t.Errorf("want %v, got %v", valueTestValue, v)
		}
	})

	mw := Value(valueTestKey, valueTestValue)

	mw(handler).ServeHTTP(w, r)
}
