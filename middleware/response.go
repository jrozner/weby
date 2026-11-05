package middleware

import (
	"net/http"

	"github.com/jrozner/weby/wrappers"
)

// WrapResponse wraps the http.ResponseWriter with a wrappers.ResponseWrapper
// that records the final response status for middleware.Logger. Register it
// before Logger. Optional capabilities remain available through
// http.ResponseController. Downstream handlers should pass the writer they
// receive directly to http.NewResponseController and keep normal response
// writes on that writer; writing through an explicitly unwrapped writer
// bypasses status tracking.
func WrapResponse(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		wrapper := wrappers.NewResponseWrapper(w)
		next.ServeHTTP(wrapper, r)
	}

	return http.HandlerFunc(fn)
}
