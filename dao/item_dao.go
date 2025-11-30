package dao

import (
	"context"
	"database/sql"
	"db/model"
	"errors"
	"fmt"
	"log"
	"time"
)

type ItemDAO interface {
	ItemInsert(ctx context.Context, item *model.Item) error
	GetItemList(ctx context.Context) ([]model.ItemSimple, error)
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
	now := time.Now()
	queryItem := `INSERT INTO items 
                  (id, user_id, name, description, price, created_at, updated_at) 
                  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, queryItem,
		item.ItemId,
		item.UserId,
		item.Name,
		item.Description,
		item.Price,
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
func (dao *itemDao) GetItemList(ctx context.Context) ([]model.ItemSimple, error) {

	query := `
		SELECT 
			i.id, 
			i.name, 
			i.price, 
			COALESCE((SELECT image_url FROM item_images WHERE item_id = i.id LIMIT 1), '') as image_url
		FROM items i
		ORDER BY i.created_at DESC`

	rows, err := dao.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fail:dao.DB.Query:%w", err)
	}
	defer rows.Close()

	// スライス（配列）の初期化
	items := make([]model.ItemSimple, 0)

	for rows.Next() {
		var i model.ItemSimple
		if err := rows.Scan(&i.ItemId, &i.Name, &i.Price, &i.ImageURL); err != nil {
			return nil, fmt.Errorf("fail:rows.Scan:%w", err)
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fail:rows.Err:%w", err)
	}

	return items, nil
}
