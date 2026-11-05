package middleware

import (
	"context"
	"net/http"
	"uuid"

	"github.com/jrozner/weby/contextvalue"
)

const RequestIDKey contextvalue.Key = "RequestID"

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID.String())
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
