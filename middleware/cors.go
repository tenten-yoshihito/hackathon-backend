package middleware

import (
	"net/http"
	"fmt"
	"strings"
	"context"
	"log"	

	"firebase.google.com/go/v4/auth"
)


type ContextKey string

const UserIDKey ContextKey = "userID"

func FirebaseAuthMiddleware(client *auth.Client, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストヘッダーから "Authorization" を取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("Authorization header missing or invalid")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// "Bearer " の後ろにあるトークン部分を抽出
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		// Firebase でトークンを検証
		token, err := client.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			log.Printf("Error verifying ID token: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// コンテキストにユーザーIDを設定
		ctx := context.WithValue(r.Context(), UserIDKey, token.UID)
		// 次のハンドラーにコンテキストを渡す
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	uid, ok := ctx.Value(UserIDKey).(string)
	if !ok || uid == "" {
		return "", fmt.Errorf("user ID not found in context")
	}
	return uid, nil
}

func CORSMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// --- 共通処理 ---
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		// ----------------
		next.ServeHTTP(w, r)
	})
}
