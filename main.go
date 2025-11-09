package main

import (
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

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// UserDBInit .envの読み取り
func UserDBInit() (*sql.DB, error) {

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("環境ファイル(.env)のロードに失敗: %w", err)
	}
	// mysqlUser := os.Getenv("MYSQL_USER")
	// mysqlUserPwd := os.Getenv("MYSQL_PASSWORD")
	// mysqlDatabase := os.Getenv("MYSQL_DATABASE")
	// dsn := fmt.Sprintf("%s:%s@(localhost:3306)/%s", mysqlUser, mysqlUserPwd, mysqlDatabase)
	// db, err := sql.Open("mysql", dsn)
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

func main() {
	// 💡 1. 環境変数 PORT を取得し、デフォルト値を設定
	port := os.Getenv("PORT")
	if port == "" {
		// 環境変数がない場合、Dockerfileや設定に合わせて8000をデフォルトとする
		port = "8000"
	}
	//--- DBの接続---
	db, err := UserDBInit()
	if err != nil {
		log.Fatalf("DBの初期化に失敗: %v", err)
	}
	log.Println("successfully connected to database")
	// --- 依存性の注入 (DI) ---
	userDAO := dao.NewUserDao(db)
	userRegister := usecase.NewUserRegister(userDAO)
	userSearch := usecase.NewUserSearch(userDAO)
	userController := controller.NewUserController(userRegister, userSearch)
	//--- 実際の処理 ---

	mux := http.NewServeMux()
	mux.HandleFunc("/user", userController.HandleUser)
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
