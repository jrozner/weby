package weby

import (
	"net/http"
	"sync"
)

type ServeMux struct {
	middleware []func(http.Handler) http.Handler
	mu         sync.RWMutex
	handler    http.Handler
	*http.ServeMux
}

func (s *ServeMux) Use(middleware func(http.Handler) http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.middleware = append(s.middleware, middleware)
	s.handler = s.buildHandler()
}

func (s *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	chain := s.handler
	s.mu.RUnlock()

	chain.ServeHTTP(w, r)
}

func (s *ServeMux) buildHandler() http.Handler {
	var chain http.Handler = s.ServeMux
	for i := len(s.middleware); i > 0; i-- {
		chain = s.middleware[i-1](chain)
	}
	return chain
}

func NewServeMux() *ServeMux {
	mux := &ServeMux{
		middleware: make([]func(http.Handler) http.Handler, 0),
		ServeMux:   http.NewServeMux(),
	}
	mux.handler = mux.ServeMux
	return mux
}
