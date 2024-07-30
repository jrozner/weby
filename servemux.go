package weby

import "net/http"

type ServeMux struct {
	middleware []func(http.Handler) http.Handler
	*http.ServeMux
}

func (s *ServeMux) Use(middleware func(http.Handler) http.Handler) {
	s.middleware = append(s.middleware, middleware)
}

func (s *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var chain http.Handler = s.ServeMux
	if len(s.middleware) > 0 {
		for i := len(s.middleware); i > 0; i-- {
			mw := s.middleware[i-1]
			chain = mw(chain)
		}
	}

	chain.ServeHTTP(w, r)
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		middleware: make([]func(http.Handler) http.Handler, 0),
		ServeMux:   http.NewServeMux(),
	}
}
