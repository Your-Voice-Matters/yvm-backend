package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// Context key type to avoid conflicts
func VerifyTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth_token := r.Header["Authorization"][0]
		if auth_token == "" {
			http.Error(w, "Authorization token missing", http.StatusUnauthorized)
			return
		}
		auth_token = auth_token[len("Bearer "):]
		secret := os.Getenv("PASSPHRASE")
		if secret == "" {
			http.Error(w, "JWT secret not set", http.StatusInternalServerError)
			return
		}
		token, err := jwt.Parse(auth_token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Store claims in request context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Failed to parse token claims", http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), "options", claims)
		next(w, r.WithContext(ctx))
	}
}
