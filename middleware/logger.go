package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jrozner/weby/wrappers"
)

func Logger(l *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			logger := l
			start := time.Now()
			if id, ok := r.Context().Value(RequestIDKey).(string); ok {
				logger = l.With("id", id)
			}

			wrapper, ok := w.(*wrappers.ResponseWrapper)
			if !ok {
				l.Error("unable to assert ResponseWriter to ResponseWrapper")
				next.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(wrapper, r)
				end := time.Now()
				logger.Info("", "status", wrapper.Status(), "remote", r.RemoteAddr, "method", r.Method, "path", r.URL.Path, "duration", end.Sub(start))
			}
		}

		return http.HandlerFunc(fn)
	}
}
