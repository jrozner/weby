package middleware

import (
	"bytes"
	"io"
	"net/http"
)

func UpdateRequestHeader(header, value string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set(header, value)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func RemoveRequestHeader(header string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del(header)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func AppendRequestBody(str string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "unable to read request body", http.StatusBadRequest)
				return
			}

			body = append(body, []byte(str)...)
			bodyBuffer := bytes.NewReader(body)

			// this is probably not the correct way to handle chunked requests
			r.ContentLength = bodyBuffer.Size()

			_ = r.Body.Close()
			r.Body = io.NopCloser(bodyBuffer)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func PrependRequest(str string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "unable to read request body", http.StatusBadRequest)
				return
			}

			body = append([]byte(str), body...)
			bodyBuffer := bytes.NewBuffer(body)
			r.ContentLength = int64(bodyBuffer.Len())

			_ = r.Body.Close()
			r.Body = io.NopCloser(bodyBuffer)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
