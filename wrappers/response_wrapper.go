package wrappers

import (
	"net/http"
)

type ResponseWrapper struct {
	status int
	http.ResponseWriter
}

func (w *ResponseWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(w.status)
}

func (w *ResponseWrapper) Status() int {
	return w.status
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		status:         0,
		ResponseWriter: w,
	}
}
