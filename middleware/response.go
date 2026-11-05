package middleware

import (
	"net/http"

	"github.com/jrozner/weby/wrappers"
)

// WrapResponse wraps the http.ResponseWriter with a wrappers.ResponseWrapper
// to allow modification of response before it is sent back to the client. This
// middleware should be used extremely early in the chain and must always be
// used before any other middleware that attempt to modify the response
func WrapResponse(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		wrapper := wrappers.NewResponseWrapper(w)
		next.ServeHTTP(wrapper, r)
	}

	return http.HandlerFunc(fn)
}
