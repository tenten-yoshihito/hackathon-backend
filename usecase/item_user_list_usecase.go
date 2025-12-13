package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type ItemUserList interface {
	GetUserItems(ctx context.Context, userID string) ([]model.ItemSimple, error)
}

type itemUserList struct {
	itemDAO dao.ItemDAO
}

// NewItemUserList : ItemUserListの生成
func NewItemUserList(itemDAO dao.ItemDAO) ItemUserList {
	return &itemUserList{itemDAO: itemDAO}
}

// GetUserItems : 指定ユーザーの出品商品一覧を取得
func (iul *itemUserList) GetUserItems(ctx context.Context, userID string) ([]model.ItemSimple, error) {
	items, err := iul.itemDAO.GetUserItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fail:itemDAO.GetUserItems:%w", err)
	}

	return items, nil
}
