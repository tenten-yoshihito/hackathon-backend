package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type ItemGet interface {
	GetItem(ctx context.Context, itemID string) (*model.Item, error)
}

type itemGet struct {
	itemDAO dao.ItemDAO
}

func NewItemGet(dao dao.ItemDAO) ItemGet {
	return &itemGet{itemDAO: dao}
}

func (us *itemGet) GetItem(ctx context.Context, itemID string) (*model.Item, error) {
	item, err := us.itemDAO.GetItem(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("fail: itemDAO.GetItem: %w", err)
	}
	return item, nil
}
