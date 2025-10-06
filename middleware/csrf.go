package middleware

import (
	"net/http"
)

func VerifyCSRFtokens(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil {
			http.Error(w, "CSRF token missing (cookie)", http.StatusForbidden)
			return
		}
		csrfHeader := r.Header.Get("X-CSRF-Token")
		if csrfHeader == "" || csrfHeader != csrfCookie.Value {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
