package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func PartnerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				http.Error(w, "partner auth not configured", http.StatusUnauthorized)
				return
			}
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < 7 || !strings.EqualFold(authHeader[:7], "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			provided := strings.TrimSpace(authHeader[7:])
			if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
