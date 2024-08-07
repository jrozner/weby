package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

const RequestIDKey = "RequestID"

func RequestID(l *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestID, err := uuid.NewUUID()
			if err != nil {
				l.Error("unable to generate request id", "error", err)
			}

			ctx := context.WithValue(r.Context(), RequestIDKey, requestID.String())
			r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
