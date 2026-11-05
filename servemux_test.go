package weby

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jrozner/weby/contextvalue"
	"github.com/jrozner/weby/middleware"
)

func TestServeMuxConstructsMiddlewareOnce(t *testing.T) {
	mux := NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	constructed := 0
	requests := 0
	mux.Use(func(next http.Handler) http.Handler {
		constructed++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			w.Header().Set("X-Request-Number", fmt.Sprint(requests))
			next.ServeHTTP(w, r)
		})
	})

	for i := 1; i <= 2; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		mux.ServeHTTP(response, request)

		if got := response.Header().Get("X-Request-Number"); got != fmt.Sprint(i) {
			t.Fatalf("request %d: X-Request-Number = %q, want %d", i, got, i)
		}
	}

	if constructed != 1 {
		t.Fatalf("middleware constructed %d times, want 1", constructed)
	}
}

func TestServeMuxRebuildsChainWhenMiddlewareIsAdded(t *testing.T) {
	mux := NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("handler"))
	})

	appendResponse := func(value string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(value))
				next.ServeHTTP(w, r)
			})
		}
	}

	mux.Use(appendResponse("first "))
	mux.Use(appendResponse("second "))

	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if got, want := response.Body.String(), "first second handler"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestServeMuxMiddlewareStillRunsForEveryRequest(t *testing.T) {
	mux := NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		id, err := contextvalue.Value[string](r.Context(), middleware.RequestIDKey)
		if err != nil {
			t.Errorf("request ID: %v", err)
			return
		}
		_, _ = w.Write([]byte(id))
	})
	mux.Use(middleware.RequestID)

	ids := make([]string, 2)
	for i := range ids {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
		ids[i] = response.Body.String()
	}

	if ids[0] == "" || ids[1] == "" {
		t.Fatalf("request IDs must be non-empty: %q", ids)
	}
	if ids[0] == ids[1] {
		t.Fatalf("request IDs must differ: %q", ids)
	}
}
