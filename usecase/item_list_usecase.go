package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type ItemList interface {
	GetItems(ctx context.Context) ([]model.ItemSimple, error)
}

type itemList struct {
	itemDAO dao.ItemDAO
}

func NewItemList(dao dao.ItemDAO) ItemList {
	return &itemList{itemDAO: dao}
}

func (us *itemList) GetItems(ctx context.Context) ([]model.ItemSimple, error) {
	
	items, err := us.itemDAO.GetItemList(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail:itemDAO.GetItemList: %w", err)
	}

	return items, nil
}