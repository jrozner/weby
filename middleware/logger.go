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
			start := time.Now()

			wrapper, ok := w.(*wrappers.ResponseWrapper)
			if !ok {
				l.Error("unable to assert ResponseWriter to ResponseWrapper")
				next.ServeHTTP(w, r)
				return
			}

			completed := false
			defer func() {
				attrs := make([]any, 0, 12)
				if completed || wrapper.Committed() {
					attrs = append(attrs, "status", wrapper.Status())
				}
				attrs = append(attrs, "remote", r.RemoteAddr, "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
				if !completed {
					attrs = append(attrs, "outcome", "panicked")
					l.ErrorContext(r.Context(), "http request", attrs...)
					return
				}

				l.InfoContext(r.Context(), "http request", attrs...)
			}()

			next.ServeHTTP(wrapper, r)
			completed = true
		}

		return http.HandlerFunc(fn)
	}
}
