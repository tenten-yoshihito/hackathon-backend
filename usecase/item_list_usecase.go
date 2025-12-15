package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
	"strings"
)

type ItemList interface {
	GetItems(ctx context.Context, limit int, offset int) ([]model.ItemSimple, error)
	SearchItems(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error)
}

type itemList struct {
	itemDAO dao.ItemDAO
}

func NewItemList(dao dao.ItemDAO) ItemList {
	return &itemList{itemDAO: dao}
}

func (us *itemList) GetItems(ctx context.Context, limit int, offset int) ([]model.ItemSimple, error) {

	items, err := us.itemDAO.GetItemList(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("fail:itemDAO.GetItemList: %w", err)
	}

	return items, nil
}

func (us *itemList) SearchItems(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error) {
	// キーワードの前後の空白を削除
	keyword = strings.TrimSpace(keyword)

	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	items, err := us.itemDAO.SearchItems(ctx, keyword, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("fail:itemDAO.SearchItems: %w", err)
	}

	return items, nil
}
