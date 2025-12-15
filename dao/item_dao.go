package dao

import (
	"context"
	"database/sql"
	"db/model"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type ItemDAO interface {
	ItemInsert(ctx context.Context, item *model.Item) error
	GetItemList(ctx context.Context, limit int, offset int) ([]model.ItemSimple, error)
	SearchItems(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error)
	GetMyItems(ctx context.Context, sellerID string) ([]model.ItemSimple, error)
	GetUserItems(ctx context.Context, userID string) ([]model.ItemSimple, error)
	GetItem(ctx context.Context, itemID string) (*model.Item, error)
	PurchaseItem(ctx context.Context, itemID string, buyerID string) error
	UpdateItem(ctx context.Context, itemID string, userID string, name string, price int, description string, embedding []float32) error
}

type itemDao struct {
	DB *sql.DB
}

// NewItemDao : ItemDAOの生成
func NewItemDao(db *sql.DB) ItemDAO {
	return &itemDao{DB: db}
}

// ItemInsert : 指定されたitemをinsertする
func (dao *itemDao) ItemInsert(ctx context.Context, item *model.Item) error {

	tx, err := dao.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("fail:txBegin(): %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("fail:tx.Rollback,%v\n", err)
		}
	}()
	var embeddingJSON interface{} = nil
	if len(item.Embedding) > 0 {
		b, err := json.Marshal(item.Embedding)
		if err == nil {
			embeddingJSON = string(b)
		}
	}

	now := time.Now()
	queryItem := `INSERT INTO items 
                  (id, user_id, name, description, price, embedding, created_at, updated_at) 
                  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, queryItem,
		item.ItemId,
		item.UserId,
		item.Name,
		item.Description,
		item.Price,
		embeddingJSON,
		now,
		now)
	if err != nil {
		return fmt.Errorf("fail:insert item: %w", err)
	}

	// 配列 (Slice) をループして、枚数分だけ INSERT を繰り返す
	queryImage := `INSERT INTO item_images (item_id, image_url, created_at) VALUES (?, ?, ?)`

	for _, imgURL := range item.ImageURLs {
		_, err := tx.ExecContext(ctx, queryImage, item.ItemId, imgURL, now)
		if err != nil {
			// 画像の保存に1枚でも失敗したら、商品登録ごと失敗させる
			return fmt.Errorf("fail:insert image: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("fail:tx.Commit(): %w", err)
	}
	return nil
}

// GetItemList : 商品一覧を取得
func (dao *itemDao) GetItemList(ctx context.Context, limit int, offset int) ([]model.ItemSimple, error) {

	query := `
		SELECT 
			i.id, 
			i.name, 
			i.price, 
			COALESCE((SELECT image_url FROM item_images WHERE item_id = i.id LIMIT 1), '') as image_url,
			i.status
		FROM items i
		ORDER BY i.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := dao.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("fail:dao.DB.Query:%w", err)
	}
	defer rows.Close()

	// スライス（配列）の初期化
	items := make([]model.ItemSimple, 0)

	for rows.Next() {
		var i model.ItemSimple
		if err := rows.Scan(&i.ItemId, &i.Name, &i.Price, &i.ImageURL, &i.Status); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return items, nil
}

// SearchItems : キーワードで商品を検索
func (dao *itemDao) SearchItems(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error) {
	query := `
		SELECT 
			i.id, 
			i.name, 
			i.price, 
			COALESCE((SELECT image_url FROM item_images WHERE item_id = i.id LIMIT 1), '') as image_url,
			i.status
		FROM items i
		WHERE i.name LIKE ?
		ORDER BY i.created_at DESC
		LIMIT ? OFFSET ?`

	// LIKE特殊文字をエスケープ
	escapedKeyword := escapeLikeString(keyword)
	// キーワードの前後に % を付けて部分一致検索
	searchKeyword := "%" + escapedKeyword + "%"

	rows, err := dao.DB.QueryContext(ctx, query, searchKeyword, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("fail:dao.DB.Query:%w", err)
	}
	defer rows.Close()

	items := make([]model.ItemSimple, 0)

	for rows.Next() {
		var i model.ItemSimple
		if err := rows.Scan(&i.ItemId, &i.Name, &i.Price, &i.ImageURL, &i.Status); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return items, nil
}

// escapeLikeString LIKE句の特殊文字をエスケープ
func escapeLikeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// GetMyItems : 自分の出品商品一覧を取得
func (dao *itemDao) GetMyItems(ctx context.Context, sellerID string) ([]model.ItemSimple, error) {
	query := `
		SELECT 
			i.id, 
			i.name, 
			i.price, 
			i.status,
			COALESCE(MIN(img.image_url), '') AS image_url
		FROM items i
		LEFT JOIN item_images img ON i.id = img.item_id
		WHERE i.user_id = ?
		GROUP BY i.id, i.name, i.price, i.status
		ORDER BY i.created_at DESC
	`

	rows, err := dao.DB.QueryContext(ctx, query, sellerID)
	if err != nil {
		return nil, fmt.Errorf("fail:db.Query:%w", err)
	}
	defer rows.Close()

	items := make([]model.ItemSimple, 0)
	for rows.Next() {
		var item model.ItemSimple
		if err := rows.Scan(&item.ItemId, &item.Name, &item.Price, &item.Status, &item.ImageURL); err != nil {
			return nil, fmt.Errorf("fail:rows.Scan:%w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return items, nil
}

// GetUserItems : 指定ユーザーの出品商品一覧を取得（プロフィールページ用）
func (dao *itemDao) GetUserItems(ctx context.Context, userID string) ([]model.ItemSimple, error) {
	query := `
		SELECT 
			i.id, 
			i.name, 
			i.price, 
			i.status,
			COALESCE(MIN(img.image_url), '') AS image_url
		FROM items i
		LEFT JOIN item_images img ON i.id = img.item_id
		WHERE i.user_id = ?
		GROUP BY i.id, i.name, i.price, i.status
		ORDER BY i.created_at DESC
	`

	rows, err := dao.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("fail:db.Query:%w", err)
	}
	defer rows.Close()

	items := make([]model.ItemSimple, 0)
	for rows.Next() {
		var item model.ItemSimple
		if err := rows.Scan(&item.ItemId, &item.Name, &item.Price, &item.Status, &item.ImageURL); err != nil {
			return nil, fmt.Errorf("fail:rows.Scan:%w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return items, nil
}

// GetItem : 指定されたitemIDの商品を取得
func (dao *itemDao) GetItem(ctx context.Context, itemID string) (*model.Item, error) {
	// 商品本体と出品者情報を取得

	// log.Printf("DEBUG: Searching item with ID: [%s]", itemID)

	queryItem := `
		SELECT 
			i.id, 
			i.user_id, 
			i.name, 
			i.price, 
			COALESCE(i.description, '') as description, 
			i.status, 
			i.created_at, 
			i.updated_at,
			u.name as seller_name,
			COALESCE(u.icon_url, '') as seller_icon_url
		FROM items i
		INNER JOIN users u ON i.user_id = u.id
		WHERE i.id = ?
	`
	row := dao.DB.QueryRowContext(ctx, queryItem, itemID)

	var item model.Item
	if err := row.Scan(
		&item.ItemId,
		&item.UserId,
		&item.Name,
		&item.Price,
		&item.Description,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.SellerName,
		&item.SellerIconURL,
	); err != nil {
		return nil, fmt.Errorf("fail: fetch item: %w", err)
	}

	// 画像一覧を取得
	queryImages := "SELECT image_url FROM item_images WHERE item_id = ?"
	rows, err := dao.DB.QueryContext(ctx, queryImages, itemID)
	if err != nil {
		return nil, fmt.Errorf("fail: fetch images: %w", err)
	}
	defer rows.Close()

	var imageURLs []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		imageURLs = append(imageURLs, url)
	}
	item.ImageURLs = imageURLs

	return &item, nil
}

// PurchaseItem : 指定されたitemIDの商品を購入済みにする
func (dao *itemDao) PurchaseItem(ctx context.Context, itemID string, buyerID string) error {

	tx, err := dao.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("fail: txBegin(): %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("fail: tx.Rollback, %v\n", err)
		}
	}()

	query := `UPDATE items SET status = ?, buyer_id = ?, purchased_at = ? WHERE id = ? AND status = ?`

	now := time.Now()
	result, err := tx.ExecContext(ctx, query, model.StatusSold, buyerID, now, itemID, model.StatusOnSale)
	if err != nil {
		return fmt.Errorf("fail: update item status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fail: get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found or already sold")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("fail: tx.Commit(): %w", err)
	}

	return nil
}

// UpdateItem : 商品情報を更新
func (dao *itemDao) UpdateItem(ctx context.Context, itemID string, userID string, name string, price int, description string, embedding []float32) error {
	tx, err := dao.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	// 商品の所有者確認と売却済みチェック
	var ownerID string
	var status string
	checkQuery := `SELECT user_id, status FROM items WHERE id = ?`
	err = tx.QueryRowContext(ctx, checkQuery, itemID).Scan(&ownerID, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("item not found")
		}
		return fmt.Errorf("failed to check item owner: %w", err)
	}

	// 所有者チェック
	if ownerID != userID {
		return model.ErrNotAuthorized
	}

	// 売却済みチェック
	if status == model.StatusSold {
		return model.ErrCannotUpdateSoldItem
	}

	// 商品情報を更新
	var embeddingJSON interface{} = nil
	if len(embedding) > 0 {
		b, err := json.Marshal(embedding)
		if err == nil {
			embeddingJSON = string(b)
		}
	}
	now := time.Now()
	updateQuery := `UPDATE items SET name = ?, price = ?, description = ?, embedding = ?, updated_at = ? WHERE id = ?`
	result, err := tx.ExecContext(ctx, updateQuery, name, price, description, embeddingJSON, now, itemID)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
