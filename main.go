package main

import (
	"context"
	"database/sql"
	"db/cache"
	"db/controller"
	"db/dao"
	"db/middleware"
	"db/service"
	"db/usecase"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// DBInit :環境変数からDB接続情報を取得し、DB接続を初期化
func DBInit() (*sql.DB, error) {

	// DB接続のための準備
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s?parseTime=true&loc=Local", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	db, err := sql.Open("mysql", connStr)

	if err != nil {
		return nil, fmt.Errorf("sql.Openで接続の確立に失敗: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("DB接続の確認(Ping)に失敗: %w", err)
	}

	return db, nil
}

// FirebaseAdminInit :Firebaseを初期化
func FirebaseAdminInit(ctx context.Context) (*auth.Client, error) {

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}
	// 認証クライアントを取得
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}
	log.Println("successfully initialized Firebase Admin SDK")
	return authClient, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("INFO: 環境ファイル(.env)のロードに失敗。Cloud Run環境を想定して続行:", err)
	}
	//  環境変数 PORT を取得し、デフォルト値を設定
	port := os.Getenv("PORT")
	if port == "" {
		// 環境変数がない場合、Dockerfileや設定に合わせて8000をデフォルトとする
		port = "8000"
	}

	// Gemini Service (Google AI Studio)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set in .env file")
	}
	geminiService := service.NewGeminiService(apiKey)

	// --- DBの接続 ---
	db, err := DBInit()
	if err != nil {
		log.Fatalf("DBの初期化に失敗: %v", err)
	}
	log.Println("successfully connected to database")
	// --- Firebase Admin SDKの初期化 ---
	authClient, err := FirebaseAdminInit(context.Background())
	if err != nil {
		log.Fatalf("Firebase Admin SDKの初期化に失敗: %v", err)
	}
	// --- 依存性の注入 (DI) ---
	// --- user ---
	userDAO := dao.NewUserDao(db)
	userRegister := usecase.NewUserRegister(userDAO)
	userSearch := usecase.NewUserSearch(userDAO)
	userGet := usecase.NewUserGet(userDAO)
	userUpdate := usecase.NewUserUpdate(userDAO)
	userController := controller.NewUserController(userRegister, userSearch, userGet, userUpdate)

	// --- notification ---
	notificationDAO := dao.NewNotificationDAO(db)

	// --- item ---
	itemDAO := dao.NewItemDao(db)
	itemRegister := usecase.NewItemRegister(itemDAO, geminiService)
	itemList := usecase.NewItemList(itemDAO)
	myItemsList := usecase.NewMyItemsList(itemDAO)
	userItemsList := usecase.NewUserItemsList(itemDAO)
	itemGet := usecase.NewItemGet(itemDAO)
	itemPurchase := usecase.NewItemPurchase(itemDAO, notificationDAO)
	itemUpdate := usecase.NewItemUpdate(itemDAO, geminiService)
	descriptionGenerate := usecase.NewDescriptionGenerate(geminiService)

	// Item controllers (refactored into 3 specialized controllers)
	itemQueryController := controller.NewItemQueryController(itemList, myItemsList, userItemsList, itemGet)
	itemCommandController := controller.NewItemCommandController(itemRegister, itemUpdate, itemPurchase)
	itemAIController := controller.NewItemAIController(descriptionGenerate)

	// --- chat ---
	chatDAO := dao.NewChatDao(db)
	chatUsecase := usecase.NewChatUsecase(chatDAO, itemDAO, notificationDAO)
	chatController := controller.NewChatController(chatUsecase)

	// --- like ---
	likeDAO := dao.NewLikeDao(db)
	likeUsecase := usecase.NewLikeUsecase(likeDAO)
	likeController := controller.NewLikeController(likeUsecase)

	// --- embedding cache (インメモリキャッシュで高速化) ---
	embeddingCache := cache.NewEmbeddingCache(itemDAO)

	// --- recommend ---
	recommendUsecase := usecase.NewRecommendUsecase(itemDAO, likeDAO, embeddingCache)
	recommendController := controller.NewRecommendController(recommendUsecase)

	// --- notification controller ---
	notificationController := controller.NewNotificationController(notificationDAO)

	// --- 実際の処理 ---

	mux := http.NewServeMux()
	// User Endpoints
	mux.HandleFunc("GET /user", userController.HandleSearchUser)
	mux.Handle("POST /register", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(userController.HandleProfileRegister)))
	mux.HandleFunc("GET /users/{id}", userController.HandleGetUser)
	mux.Handle("PUT /users/me", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(userController.HandleUpdateUser)))

	// Item Query Endpoints
	mux.HandleFunc("GET /items", itemQueryController.HandleItemList)
	mux.HandleFunc("GET /items/{id}", itemQueryController.HandleItemDetail)
	mux.Handle("GET /items/my", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(itemQueryController.HandleMyItems)))
	mux.HandleFunc("GET /users/{userId}/items", itemQueryController.HandleUserItems)
	mux.HandleFunc("GET /items/{id}/recommend", recommendController.HandleGetRecommendations)
	mux.Handle("GET /items/recommend", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(recommendController.HandleGetPersonalizedRecommendations)))

	// 商品出品 (POST /items)
	mux.Handle("POST /items", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(itemCommandController.HandleItemRegister)))
	// 商品購入 (POST /items/{id}/purchase)
	mux.Handle("POST /items/{id}/purchase", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(itemCommandController.HandleItemPurchase)))
	// 商品更新 (PUT /items/{id})
	mux.Handle("PUT /items/{id}", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(itemCommandController.HandleItemUpdate)))
	// AI商品説明生成 (POST /items/generate-description)
	mux.Handle("POST /items/generate-description", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(itemAIController.HandleGenerateDescription)))

	// Like Endpoints
	mux.Handle("POST /items/{id}/like", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(likeController.HandleToggleLike)))
	mux.Handle("GET /items/liked", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(likeController.HandleGetLikedItems)))
	mux.Handle("GET /items/liked-ids", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(likeController.HandleGetLikedItemIDs)))

	// Chat Endpoints
	mux.Handle("POST /items/{item_id}/chat", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(chatController.HandleGetOrCreateRoom)))
	mux.Handle("GET /items/{item_id}/chat_rooms", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(chatController.HandleGetChatRoomList)))
	mux.Handle("GET /chats/{room_id}/messages", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(chatController.HandleGetMessages)))
	mux.Handle("POST /chats/{room_id}/messages", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(chatController.HandleSendMessage)))

	// Notification Endpoints
	mux.Handle("GET /notifications", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(notificationController.HandleGetNotifications)))
	mux.Handle("GET /notifications/unread", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(notificationController.HandleGetUnreadCount)))
	mux.Handle("PUT /notifications/{id}/read", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(notificationController.HandleMarkAsRead)))
	mux.Handle("PUT /notifications/read-all", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(notificationController.HandleMarkAllAsRead)))

	// CORS Middlewareを適用
	wrappedHandler := middleware.CORSMiddleware(mux)

	closeDBWithSysCall(db)

	addr := ":" + port
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, wrappedHandler); err != nil {
		log.Fatal(err)
	}
}

// closeDBWithSysCall :Ctrl+CでHTTPサーバー停止時にDBをクローズ
func closeDBWithSysCall(db *sql.DB) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)

		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("success: db.Close()")
		os.Exit(0)
	}()
}
