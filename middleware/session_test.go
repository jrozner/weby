package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
)

/*
tests:
- valid session
- invalid session
- no session
*/

const (
	defaultCookieName = "SESSION"
)

func setupSessionMiddleware(cookieName string, store *sessions.CookieStore) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return Session(cookieName, store)(handler)
}

func TestSessionValid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	store := sessions.NewCookieStore([]byte("test"))

	session, err := store.Get(r, defaultCookieName)
	if err != nil {
		t.Errorf("unable to get initial session: %s", err)
	}

	err = session.Save(r, w)
	if err != nil {
		t.Errorf("unable to save session: %s", err)
	}

	savedCookie, err := http.ParseSetCookie(w.Header().Get("Set-Cookie"))
	if err != nil {
		t.Errorf("unable to parse Set-Cookie: %s", err)
	}

	mw := setupSessionMiddleware(defaultCookieName, store)
	cookie := http.Cookie{
		Name:  defaultCookieName,
		Value: savedCookie.Value,
	}

	r = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()
	r.AddCookie(&cookie)
	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("request failed with valid session")
	}
}

func TestSessionInvalid(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))

	mw := setupSessionMiddleware(defaultCookieName, store)
	cookie := &http.Cookie{
		Name:  defaultCookieName,
		Value: "MTczMTg2NjA5N3xEWDhFQVFMX2dBQUJFQUVRQUFELUJfWF9nQUFFQm5OMGNtbHVad3dOQUF0QlkyTmxjM05VYjJ0bGJnWnpkSEpwYm1jTV9nT1VBUDREa0dWNVNuSmhWMUZwVDJsS2JFOUlRazFpUmxaclVWaEtWMUZWT1daaVZqaDNWR3BLTTFwWVp6Qk9TSEF6VkRCck1tSnFUWGhWVkdRMFlXeHdVVnByYURCUFdFMDBTV2wzYVdSSWJIZEphbTlwV1ZoQ2QySkhiR3BaV0ZKd1lqSTFZMHd5T1hKa1IwVjBZVmMxTUZwWVNuVlpWM2QwV1ZoUmNtRnVaREJKYVhkcFdWZDRia2xxYjJsVmJFMTVUbFJaYVdaUkxtVjVTakphV0VscFQycEZjMGx0Y0RCaFUwazJTV3RHVlV4dGFHbGlWa1pQVVZWb2VsSlhkRkpTYmsweFZWVmFObFF3UmtSaE0yeGhZVEJLUzFoNU1YRmxiWGh0VlRGQ00xZEZSbGhoVkVwbVZFVkZhVXhEU25Cak0wMXBUMmxLYjJSSVVuZGplbTkyVERKU2JHUnBNSGxPYWswd1RrUkplVTE1TlhaaE0xSm9URzFPZG1KVFNYTkpiVVl4V2tOSk5rbHRhREJrU0VKNlQyazRkbHBIVmpKTVZFa3lUWHBSTUUxcVNYcE1iVGx5WkVkRmRWa3lPWFJKYVhkcFl6TldhVWxxYjJsaGJUbHNVVWRTYkZsWFVtbGxXRkpzWTNrMWRWcFlVV2xNUTBwd1dWaFJhVTlxUlROTmVrVTBUbXBaZDA5VVZYTkpiVlkwWTBOSk5rMVVZM3BOVkdjeVQxUlpOVTVUZDJsWk1teHJTV3B2YVUxSE9XaGhNMEkyV1ZoYWJtSkZjRWhpUjJ4UlpHMTNNVnBFWTJsTVEwb3hZVmRSYVU5cFNYZE5TRlp5WTBkV2MwNUhPRE5rTVVwNFkxZE9hRTVFVm10T2VVbHpTVzVPYW1ORFNUWlhlVXAyWTBkV2RXRlhVV2xNUTBwM1kyMDViV0ZYZUd4SmFYZHBXbGN4YUdGWGQybFlVM2RwV1ZoV01HRkdPVEJoVnpGc1NXcHZlRTU2VFhoUFJGa3lUVVJyTUdaUkxrdERXV0pNUldGTFQxVlFjRTFrUjBJd2F6aExjV2h0TUZReldYRnFXalI0YWpkQ2RsSnRRblJtZURSSVNHMVBTVkJNZUVKRU5uUk1WMGxOYXpCNFRtUnBabmxyTkRacldVdzVRVE0wYVRoVFltOTNURGxSU25RMmVsWm9RbTEwVTJwMU1HeGhaR1ZYU1RKUWF6bHliRGhVWjBjNWVscHlSazVYTkc5MmJrSlJVMVoxVTFSR1ltRlZhMVp3YVVJNVlXdHVXa0pFUlhSQ2RGcEVUVVpHTUVkbWIxZENVMFppWTBoSlFrdFNVMWhaVGxWYVNsRlVTVEp0Y25WeE0yTlNZbUk1TUhRMmVGSTJTeTA1TjA4d1pVTjBjVE5rVlZJeFRXSXlORFY2U0hGeVdrczBaR3BKUmpOb1RrWlBMWG81V0RJM01ucE9aMDlmVDBKNmJYcDNObVUxYjBsS2FVeGpRekV0V1hCRFkxRlZjRWMxVDJKV0xWQTVSelJVT1dWNmJ6STVOR2RJVURGWUxXbFZUblpRVG1ObmRWSkdRemx5ZFdWS1luQnFSRUl0Umt0UE1qRk1ibWREUTFJNWNFOTRkVkJqVkdod1VRWnpkSEpwYm1jTUNRQUhTVVJVYjJ0bGJnWnpkSEpwYm1jTV9nUGxBUDRENFdWNVNuSmhWMUZwVDJsS2JFOUlRazFpUmxaclVWaEtWMUZWT1daaVZqaDNWR3BLTTFwWVp6Qk9TSEF6VkRCck1tSnFUWGhWVkdRMFlXeHdVVnByYURCUFdFMDBTV2wzYVZsWGVHNUphbTlwVld4TmVVNVVXV2xtVVM1bGVVcDZaRmRKYVU5cFNYZE5TRlp5WTBkV2MwNUhPRE5rTVVwNFkxZE9hRTVFVm10T2VVbHpTVzAxYUdKWFZXbFBhVXBMWWpKVloxVnRPVFppYlZaNVNXbDNhVnBYTVdoaFYzZHBUMmxLY1dJeVZrRmFSMVpvV2tkS05XUkhWbnBNYlRWc1pFTkpjMGx1V214amFVazJUVk4zYVdGWVRucEphbTlwWVVoU01HTklUVFpNZVRscldsaFpkRTFxV1hwT1JGRjVUV3BOZFdJeWREQlpVelZxWWpJd2FVeERTbWhrVjFGcFQybEpkMkl5Um5KalNIQm9aRzFrYzFOclpITmhWa0l5WWtSV2EwNTVTWE5KYld4b1pFTkpOazFVWTNwTlZHY3lUbXBCTlU1VGQybGFXR2gzU1dwdmVFNTZUWGhQUkZrMVRtcHJNVXhEU25Ga1IydHBUMmxLU2xKRE5XMVVXR3Q2VWpKNFZrMHpaRXhVVkU1d1lsWnNjazB6UW5SVk1XUXdWMFU1YVdWRVFYbFNlVEZ4VEZaR2FFNXJSbEJZTWxKcldXeE9Ta2xwZDJsWlZ6RjVTV3B3WWtsdVFqTmFRMHBrVEVOS2NGcElRV2xQYVVsM1RVYzVjbU5IVm5OT1Iyc3hZVmhGTUZadE9URldSRlpyVG5sSmMwbHROWFppYlU1c1NXcHZhVmt3Y0RCVE0wWjVaSHBrYW1GdFpIaFVNSEJTWldwQ1MxcFZlRU5hZWpBNVNXbDNhV05JU214YWJWWjVZMjFXYTFnelZucGFXRXAxV1ZjeGJFbHFiMmxoYlRsc1VVZFNiRmxYVW1sbFdGSnNZM2sxZFZwWVVXbE1RMHBvWkZoU2IxZ3pVbkJpVjFWcFQycEZNMDE2UlRST2FsbDNUMVJSYzBsdFJqQllNbWhvWXpKbmFVOXBTa3RPTVdnd1pFZEtRMkZZVGxWU2VrSlNWMFZ3ZFZsc1JYaFJNblJ1U1c0d0xuQTBPVTV6ZHkwd2NtdFlNekJOU0RWWFNIZDZlVk5uTFZoblVGOUNlWEZsYW14dWIycFVORGhGTW1aS1l6Wk1Xa1o0VTJwTGJsQTNVRWd4YzBSYVJuSXRaVnB4UWpkaVMwWjFSakpKTUU4eFl6bEdhR3RFUTBGVkxYbFJTRmxyV204eFVWaEtiMVJXUnpneGVsVXpTVnB2TWtSWGNITlVObVpEWkZVdGNrdE1OVXM1UTBwTWVXcGpSbEpQTmxSb2RqTlpTRVpWTFVaeGMwMUZTemxEUzBOUlYyc3RZbmg2U1ZGV1UwWnBhelZGUkhwSFZUWlFiRjl2UzAxNFVGaFFUSEZvYkRjM1RscDBNbE5JVG5WblYxQkdiWEZaVlRWaE1qSlZYMXBvTWpkbmRVaEhiV3BQTUhOTGFuWnBhV0ZoZWxkc2FtaFJNM0pYV2pRNWVtWm1jWGxtTUVOdFJGZEhSWFJsUWtKbFdHOW1kQzF3VWtwWFNWaFhaa1psU1hkeFptRmpiRWw0UzFKa2IwaHRhbTk2UzJSNGRFUm5kRzF5VmxsUmJUQlhkbE53YnpZeE1WQTNabXROTmxVMGNIRXlVVU5xV0ZsSWR3WnpkSEpwYm1jTURBQUtRM1Z6ZEc5dFpYSkpSQVZwYm5Rek1nUUNBQUlHYzNSeWFXNW5EQWdBQmxWelpYSkpSQVZwYm5Rek1nUUNBQUk9fHHXBS8wMYrroaO0I3xm3Ywy6a59H1BzbY5cPWYZbRXa",
	}

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.AddCookie(cookie)
	mw.ServeHTTP(w, r)

	setCookie := w.Result().Header.Get("Set-Cookie")
	cookie, err := http.ParseSetCookie(setCookie)
	if err != nil {
		t.Errorf("no session cookie returned")
	}

	if cookie.Name != defaultCookieName {
		t.Errorf("returned cookie is not a session cookie")
	}
}

func TestSessionNone(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))

	mw := setupSessionMiddleware(defaultCookieName, store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)

	setCookie := w.Result().Header.Get("Set-Cookie")
	cookie, err := http.ParseSetCookie(setCookie)
	if err != nil {
		t.Errorf("no session cookie returned")
	}

	if cookie.Name != defaultCookieName {
		t.Errorf("returned cookie is not a session cookie")
	}
}

func TestInvalidCookieName(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))

	mw := setupSessionMiddleware("", store)

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status code of %d but got %d", http.StatusInternalServerError, w.Result().StatusCode)
	}
}
