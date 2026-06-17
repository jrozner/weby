package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/jrozner/weby"
)

/*
tests
- valid csrf token, unsafe method
- invalid csrf token (not base64), unsafe method
- invalid csrf token (wrong size), unsafe method
- invalid csrf token (mismatch), unsafe method
- invalid csrf token safe method
- no csrf token, unsafe method
- no csrf token, safe method
*/

func setupCSRFMiddleware(cookieName string, store *sessions.CookieStore) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux := weby.NewServeMux()
	mux.Use(Session(cookieName, store))
	mux.Use(CSRF(cookieName, store))
	mux.Handle("/test", handler)

	return mux
}

func TestValidTokenUnsafeMethod(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	mw := setupCSRFMiddleware(defaultCookieName, store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/test", nil)

	cookies := w.Result().Cookies()
	if len(cookies) < 1 {
		t.Error("no cookies found")
	}

	// multiple calls to Save on the session will result in multiple cookie headers for the session cookie. In a
	// browser, the last one will win. That's the behavior we want here
	last := cookies[len(cookies)-1]
	r.AddCookie(last)

	r.Header.Add("X-CSRF-Token", w.Result().Header.Get("X-CSRF-Token"))

	w = httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("request failed with valid session")
	}
}

func TestNotBase64TokenUnsafeMethod(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	mw := setupCSRFMiddleware(defaultCookieName, store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/test", nil)

	cookies := w.Result().Cookies()
	if len(cookies) < 1 {
		t.Error("no cookies found")
	}

	// multiple calls to Save on the session will result in multiple cookie headers for the session cookie. In a
	// browser, the last one will win. That's the behavior we want here
	last := cookies[len(cookies)-1]
	r.AddCookie(last)

	r.Header.Add("X-CSRF-Token", "~~~BADTOKEN~~~")

	w = httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("malformed token (non-base64) resulted in non-bad request")
	}
}

func TestWrongSizeTokenUnsafeMethod(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	mw := setupCSRFMiddleware(defaultCookieName, store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/test", nil)

	cookies := w.Result().Cookies()
	if len(cookies) < 1 {
		t.Error("no cookies found")
	}

	// multiple calls to Save on the session will result in multiple cookie headers for the session cookie. In a
	// browser, the last one will win. That's the behavior we want here
	last := cookies[len(cookies)-1]
	r.AddCookie(last)

	r.Header.Add("X-CSRF-Token", "dGVzdA==")

	w = httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("malformed token (wrong size) resulted in non-bad request")
	}
}

func TestMismatchTokenUnsafeMethod(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	mw := setupCSRFMiddleware(defaultCookieName, store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/test", nil)

	cookies := w.Result().Cookies()
	if len(cookies) < 1 {
		t.Error("no cookies found")
	}

	// multiple calls to Save on the session will result in multiple cookie headers for the session cookie. In a
	// browser, the last one will win. That's the behavior we want here
	last := cookies[len(cookies)-1]
	r.AddCookie(last)

	r.Header.Add("X-CSRF-Token", "sQoZUhVzkCdcXZovQO8SpW052osNlrl0pWCNusT58A2viDdt8lPU7GwGEmYse3mvTA+DstlfyILYKS6yBMlGEQ==")

	w = httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("malformed token (mismatched token) resulted in non-bad request")
	}
}

func TestInvalidTokenSafeMethod(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	mw := setupCSRFMiddleware(defaultCookieName, store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodGet, "/test", nil)

	cookies := w.Result().Cookies()
	if len(cookies) < 1 {
		t.Error("no cookies found")
	}

	// multiple calls to Save on the session will result in multiple cookie headers for the session cookie. In a
	// browser, the last one will win. That's the behavior we want here
	last := cookies[len(cookies)-1]
	r.AddCookie(last)

	r.Header.Add("X-CSRF-Token", "sQoZUhVzkCdcXZovQO8SpW052osNlrl0pWCNusT58A2viDdt8lPU7GwGEmYse3mvTA+DstlfyILYKS6yBMlGEQ==")

	w = httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("invalid token with safe method resulted in non-ok request")
	}
}

func TestNoCSRFTokenSafeMethod(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	mw := setupCSRFMiddleware(defaultCookieName, store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodGet, "/test", nil)

	cookies := w.Result().Cookies()
	if len(cookies) < 1 {
		t.Error("no cookies found")
	}

	// multiple calls to Save on the session will result in multiple cookie headers for the session cookie. In a
	// browser, the last one will win. That's the behavior we want here
	last := cookies[len(cookies)-1]
	r.AddCookie(last)

	w = httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("missing token with safe method resulted in non-ok request")
	}
}

func TestNoCSRFTokenUnsafeMethod(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	mw := setupCSRFMiddleware(defaultCookieName, store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	r = httptest.NewRequest(http.MethodPost, "/test", nil)

	cookies := w.Result().Cookies()
	if len(cookies) < 1 {
		t.Error("no cookies found")
	}

	// multiple calls to Save on the session will result in multiple cookie headers for the session cookie. In a
	// browser, the last one will win. That's the behavior we want here
	last := cookies[len(cookies)-1]
	r.AddCookie(last)

	w = httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("missing token unsafe method resulted in non-bad request")
	}
}
