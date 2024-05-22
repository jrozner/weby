package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/jrozner/weby/wrappers"
)

func Logger(l *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapper, ok := w.(*wrappers.ResponseWrapper)
			if !ok {
				l.Println("unable to assert ResponseWriter to ResponseWrapper")
				next.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(wrapper, r)
				end := time.Now()
				l.Printf("%s %s: %s %d %dms\n", time.Now().Format(time.RFC3339), r.RemoteAddr, r.URL.Path, wrapper.StatusCode, end.Sub(start)/time.Millisecond)
			}
		}

		return http.HandlerFunc(fn)
	}
}
