package dao

import (
	"context"
	"database/sql"
	"db/model"
	"fmt"
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

func (dao *itemDao) DBInsert(ctx context.Context, item *model.Item) error {

	tx, err := dao.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("fail:txBegin(): %w", err)
	}

	query := "INSERT INTO items (id, name, description, price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = tx.ExecContext(ctx, query, item.ItemId, item.Name, item.Description, item.Price, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("fail:tx.ExecContext(): %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("fail:tx.Commit(): %w", err)
	}

	return nil
}
