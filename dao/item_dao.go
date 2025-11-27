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
}

type itemDao struct {
	DB *sql.DB
}

func NewItemDao(db *sql.DB) ItemDAO {
	return &itemDao{DB: db}
}

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
