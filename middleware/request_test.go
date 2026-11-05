package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestBodyMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		middleware func(http.Handler) http.Handler
		want       string
	}{
		{
			name:       "append",
			middleware: AppendRequestBody(" suffix"),
			want:       "body suffix",
		},
		{
			name:       "prepend",
			middleware: PrependRequest("prefix "),
			want:       "prefix body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			var contentLength int64
			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read request body: %v", err)
				}
				got = string(body)
				contentLength = r.ContentLength
			})

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("body"))
			tt.middleware(next).ServeHTTP(httptest.NewRecorder(), req)

			if got != tt.want {
				t.Fatalf("body = %q, want %q", got, tt.want)
			}
			if contentLength != int64(len(tt.want)) {
				t.Fatalf("content length = %d, want %d", contentLength, len(tt.want))
			}
		})
	}
}

func TestRequestBodyMiddlewareReadError(t *testing.T) {
	middleware := []struct {
		name string
		wrap func(http.Handler) http.Handler
	}{
		{name: "append", wrap: AppendRequestBody(" suffix")},
		{name: "prepend", wrap: PrependRequest("prefix ")},
	}

	for _, tt := range middleware {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			})
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Body = errorReadCloser{}
			response := httptest.NewRecorder()

			tt.wrap(next).ServeHTTP(response, req)

			if called {
				t.Fatal("next handler called after request body read error")
			}
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

type errorReadCloser struct{}

func (errorReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("read error")
}

func (errorReadCloser) Close() error {
	return nil
}
