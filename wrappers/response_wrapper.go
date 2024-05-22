package wrappers

import (
	"bytes"
	"net/http"
)

type ResponseWrapper struct {
	StatusCode int
	body       *bytes.Buffer
	http.ResponseWriter
}

func (w *ResponseWrapper) WriteHeader(code int) {
	w.StatusCode = code
}

func (w *ResponseWrapper) Write(body []byte) (int, error) {
	return w.body.Write(body)
}

func (w *ResponseWrapper) Close() error {
	w.ResponseWriter.WriteHeader(w.StatusCode)
	_, err := w.ResponseWriter.Write(w.body.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		StatusCode:     0,
		body:           bytes.NewBuffer(make([]byte, 0, 4096)),
		ResponseWriter: w,
	}
}
