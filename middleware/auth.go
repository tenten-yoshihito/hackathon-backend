package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

type contextKey string

const userIDKey contextKey = "userID"

func FirebaseAuthMiddleware(client *auth.Client, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == "" || idToken == authHeader { // Bearerがない、または空
			log.Println("auth: invalid header format")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := client.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			log.Printf("auth: verification failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	uid, ok := ctx.Value(userIDKey).(string)
	if !ok || uid == "" {
		return "", fmt.Errorf("uid not found in context")
	}
	return uid, nil
}
