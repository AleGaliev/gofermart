package middleware

import (
	"net/http"
)

type JWTManager interface {
	GetLoginFromToken(tokenString string) (string, error)
}

func AuthMiddleware(jwtManager JWTManager) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("Authorization")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			user, err := jwtManager.GetLoginFromToken(c.Value)

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-User-Login", user)

			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
