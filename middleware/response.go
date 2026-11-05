package middleware

import (
	"net/http"

	"github.com/jrozner/weby/wrappers"
)

// WrapResponse wraps the http.ResponseWriter with a wrappers.ResponseWrapper
// that records the final response status for middleware.Logger. Register it
// before Logger. Optional capabilities remain available through
// http.ResponseController.
func WrapResponse(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		wrapper := wrappers.NewResponseWrapper(w)
		next.ServeHTTP(wrapper, r)
	}

	return http.HandlerFunc(fn)
}
