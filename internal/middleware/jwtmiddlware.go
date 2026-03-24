package middleware

import (
	"FCH/internal/jwt"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type contextkey string

const (
	userIDKey   contextkey = "userID"
	usernameKey contextkey = "username"
)

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		tokenString := cookie.Value
		if tokenString == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:    "jwt",
				Value:   "",
				Expires: time.Now().Add(-1 * time.Hour),
				Path:    "/",
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), usernameKey, claims.UserID)
		ctx = context.WithValue(ctx, usernameKey, claims.Username)

		log.Printf("userID from middleware: %d, userID from middleware: %s", claims.UserID, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromRequest(r *http.Request) (uint, error) {
	if userIDVal := r.Context().Value(userIDKey); userIDVal != nil {
		if userID, ok := userIDVal.(uint); ok {
			return userID, nil
		}
	}
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return 0, fmt.Errorf("error get cookie from request: %w", err)
	}
	log.Printf("cookie from jwt: %s", cookie)
	token, err := jwt.ValidateToken(cookie.Value)
	if err != nil {
		return 0, fmt.Errorf("error get id from token: %w", err)
	}
	log.Printf("userID from jwt: %d", token.UserID)
	return token.UserID, nil

}
