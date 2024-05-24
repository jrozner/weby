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

func UpdateResponseHeader(header, value string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			wrapper, ok := w.(*wrappers.ResponseWrapper)
			if !ok {
				next.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(wrapper, r)
				wrapper.Header().Set(header, value)
			}
		}

		return http.HandlerFunc(fn)
	}
}

func RemoveResponseHeader(header string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			wrapper, ok := w.(*wrappers.ResponseWrapper)
			if !ok {
				next.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(wrapper, r)
				wrapper.Header().Del(header)
			}
		}

		return http.HandlerFunc(fn)
	}
}
