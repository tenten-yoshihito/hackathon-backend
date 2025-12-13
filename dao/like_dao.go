package dao

import (
	"context"
	"database/sql"
	"db/model"
	"fmt"
	"time"
)

type LikeDAO interface {
	ToggleLike(ctx context.Context, userID, itemID string) error
	GetLikedItems(ctx context.Context, userID string) ([]model.ItemSimple, error)
	GetLikedItemIDs(ctx context.Context, userID string) ([]string, error)
}

type likeDao struct {
	db *sql.DB
}

func NewLikeDao(db *sql.DB) LikeDAO {
	return &likeDao{db: db}
}

// ToggleLike: 既にいいね済の場合: DELETE 文を実行して「いいね解除」/まだいいねしていない場合: INSERT 文を実行して「いいね登録」
func (dao *likeDao) ToggleLike(ctx context.Context, userID, itemID string) error {
	// Check if like already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND item_id = ?)`
	err := dao.db.QueryRowContext(ctx, checkQuery, userID, itemID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check like existence: %w", err)
	}

	if exists {
		// Delete the like
		deleteQuery := `DELETE FROM likes WHERE user_id = ? AND item_id = ?`
		_, err = dao.db.ExecContext(ctx, deleteQuery, userID, itemID)
		if err != nil {
			return fmt.Errorf("failed to delete like: %w", err)
		}
	} else {
		// Insert the like
		insertQuery := `INSERT INTO likes (user_id, item_id, created_at) VALUES (?, ?, ?)`
		_, err = dao.db.ExecContext(ctx, insertQuery, userID, itemID, time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert like: %w", err)
		}
	}

	return nil
}

// GetLikedItems 「マイページの『いいねした商品』タブ」 で表示するためのデータ取得メソッド
func (dao *likeDao) GetLikedItems(ctx context.Context, userID string) ([]model.ItemSimple, error) {
	query := `
		SELECT 
			i.id, 
			i.name, 
			i.price, 
			i.status,
			COALESCE(MIN(img.image_url), '') AS image_url,
			MAX(l.created_at) AS liked_at
		FROM likes l
		INNER JOIN items i ON l.item_id = i.id
		LEFT JOIN item_images img ON i.id = img.item_id
		WHERE l.user_id = ?
		GROUP BY i.id, i.name, i.price, i.status
		ORDER BY liked_at DESC
	`

	rows, err := dao.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query liked items: %w", err)
	}
	defer rows.Close()

	items := make([]model.ItemSimple, 0)
	for rows.Next() {
		var item model.ItemSimple
		var likedAt time.Time
		if err := rows.Scan(&item.ItemId, &item.Name, &item.Price, &item.Status, &item.ImageURL, &likedAt); err != nil {
			return nil, fmt.Errorf("failed to scan liked item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return items, nil
}

// GetLikedItemIDs : 「商品一覧画面で、自分がいいね済の商品だけハートを赤くする」 ためのメソッド
func (dao *likeDao) GetLikedItemIDs(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT item_id FROM likes WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := dao.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query liked item IDs: %w", err)
	}
	defer rows.Close()

	itemIDs := make([]string, 0)
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err != nil {
			return nil, fmt.Errorf("failed to scan item ID: %w", err)
		}
		itemIDs = append(itemIDs, itemID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return itemIDs, nil
}
