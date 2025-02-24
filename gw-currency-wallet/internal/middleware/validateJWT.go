package middleware

import (
	"context"

	"net/http"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("your_secret_key")

const UserNameconst = "username"
const User_id = "user_id"

type Claims struct {
	Username string `json:"username"`
	Id       int    `json:"id"`
	jwt.StandardClaims
}

func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenStr = tokenStr[len("Bearer "):]
		claims := new(Claims)
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, UserNameconst, claims.Username)
		ctx = context.WithValue(ctx, User_id, claims.Id)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
