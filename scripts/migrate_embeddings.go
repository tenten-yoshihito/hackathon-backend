// scripts/migrate_embeddings.go
// 既存商品にベクトルを一括生成するスクリプト

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	// .envファイルを読み込む
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// DB接続
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s?parseTime=true&loc=Local", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Gemini API初期化
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	// embeddingがnullの商品を取得
	query := `SELECT id, name, COALESCE(description, '') as description FROM items WHERE embedding IS NULL`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Fatalf("Failed to query items: %v", err)
	}
	defer rows.Close()

	type Item struct {
		ID          string
		Name        string
		Description string
	}

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		items = append(items, item)
	}

	log.Printf("Found %d items without embeddings", len(items))

	// 各商品にベクトルを生成
	successCount := 0
	failCount := 0

	for i, item := range items {
		log.Printf("[%d/%d] Processing item: %s (ID: %s)", i+1, len(items), item.Name, item.ID)

		// テキストを結合
		textToEmbed := fmt.Sprintf("%s\n%s", item.Name, item.Description)

		// ベクトル生成
		res, err := client.Models.EmbedContent(ctx, "models/text-embedding-004",
			[]*genai.Content{
				{
					Parts: []*genai.Part{
						genai.NewPartFromText(textToEmbed),
					},
				},
			},
			nil,
		)
		if err != nil {
			log.Printf("  ❌ Failed to generate embedding: %v", err)
			failCount++
			continue
		}

		if res == nil || len(res.Embeddings) == 0 || len(res.Embeddings[0].Values) == 0 {
			log.Printf("  ❌ Empty embedding returned")
			failCount++
			continue
		}

		embedding := res.Embeddings[0].Values

		// JSONに変換
		embeddingJSON, err := json.Marshal(embedding)
		if err != nil {
			log.Printf("  ❌ Failed to marshal embedding: %v", err)
			failCount++
			continue
		}

		// DBに更新
		updateQuery := `UPDATE items SET embedding = ? WHERE id = ?`
		_, err = db.ExecContext(ctx, updateQuery, string(embeddingJSON), item.ID)
		if err != nil {
			log.Printf("  ❌ Failed to update database: %v", err)
			failCount++
			continue
		}

		log.Printf("  ✅ Successfully updated")
		successCount++
	}

	log.Printf("\n=== Migration Complete ===")
	log.Printf("Success: %d items", successCount)
	log.Printf("Failed: %d items", failCount)
	log.Printf("Total: %d items", len(items))
}
