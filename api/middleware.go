package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

// Middleware to handle login auth.
func (apiHandler *APIHandler) AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth, ok := r.Header["Authorization"]
		if len(auth) < 1 && !ok {
			apiHandler.logger.Error("No auth header.")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

		tokenString := strings.TrimPrefix(auth[0], "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Type assertion.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				apiHandler.logger.Errorf("Signing method not right: %v\n", token.Header["alg"])
				return nil, fmt.Errorf("Unexpected signing method: %v\n", token.Header["alg"])
			}
			return []byte(apiHandler.signingKey), nil
		})

		if nil != err {
			apiHandler.logger.Error("JWT token parse error.")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			email, ok := claims["email"].(string)
			if ok {
				// Add email to request's context so that can get account by email.
				ctx := context.WithValue(r.Context(), "email", email)
				h(w, r.WithContext(ctx))
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
	}
}
