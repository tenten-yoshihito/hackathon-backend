package main

import (
	"context"
	"database/sql"
	"db/controller"
	"db/dao"
	"db/middleware"
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

// UserDBInit,環境変数からDB接続情報を取得し、DB接続を初期化する
func DBInit() (*sql.DB, error) {

	// DB接続のための準備
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
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

// Firebaseの初期化
func FirebaseAdminInit(ctx context.Context) (*auth.Client, error) {

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}
	// 認証クライアントの取得
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
	//--- DBの接続---
	db, err := DBInit()
	if err != nil {
		log.Fatalf("DBの初期化に失敗: %v", err)
	}
	log.Println("successfully connected to database")
	//--- Firebase Admin SDKの初期化 ---
	authClient, err := FirebaseAdminInit(context.Background())
	if err != nil {
		log.Fatalf("Firebase Admin SDKの初期化に失敗: %v", err)
	}
	// --- 依存性の注入 (DI) ---
	userDAO := dao.NewUserDao(db)
	userRegister := usecase.NewUserRegister(userDAO)
	userSearch := usecase.NewUserSearch(userDAO)
	userController := controller.NewUserController(userRegister, userSearch)

	itemDAO := dao.NewItemDao(db)
	itemRegister := usecase.NewItemRegister(itemDAO)
	itemList := usecase.NewItemList(itemDAO)
	itemGet := usecase.NewItemGet(itemDAO)
	itemController := controller.NewItemController(itemRegister, itemList, itemGet)
	//--- 実際の処理 ---

	mux := http.NewServeMux()
	// User Endpoints
	mux.HandleFunc("/user", userController.HandleSearchUser)
	mux.Handle("/register", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(userController.HandleProfileRegister)))

	// Item Endpoints
	// 商品一覧 (GET /items)
	mux.HandleFunc("GET /items", itemController.HandleItemList)
	// 商品出品 (POST /items)
	mux.Handle("POST /items", middleware.FirebaseAuthMiddleware(authClient, http.HandlerFunc(itemController.HandleItemRegister)))
	// 商品詳細 (GET /items/{itemID})
	mux.HandleFunc("GET /items/", itemController.HandleItemDetail)

	// CORS Middlewareを適用
	wrappedHandler := middleware.CORSMiddleware(mux)

	closeDBWithSysCall(db)

	addr := ":" + port
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, wrappedHandler); err != nil {
		log.Fatal(err)
	}
}

// Ctrl+CでHTTPサーバー停止時にDBをクローズする
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
