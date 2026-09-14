package middleware

import (
	"context"
	"net/http"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

// AuthApi intercepts the request and checks the token
func AuthApi(jwtKey []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Public route
		if r.URL.Path == "/user/login" || r.URL.Path == "/ping" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("authToken")
		if err != nil {
			// If there is no cookie, they are not logged in
			http.Error(w, "Unauthorized: Missing cookie", http.StatusUnauthorized)
			return
		}

		// Extract the raw JWT string from the cookie
		tokenString := cookie.Value

		// Parse and validate the token using the injected key
		claims := &auth.Claim{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Invalid Token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.AuthUserIdKey, claims.UserId)

		// If valid, let the request pass through to the Handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
