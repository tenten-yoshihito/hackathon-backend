package usecase

import (
	"context"
	"db/dao"
	"db/model"
)

type ItemDetail interface {
	GetItem(ctx context.Context, itemId string) (*model.Item, error)
}

type itemDetail struct {
	dao dao.ItemDAO
}

func NewItemDetail(dao dao.ItemDAO) ItemDetail {
	return &itemDetail{dao: dao}
}

func (u *itemDetail) GetItem(ctx context.Context, itemId string) (*model.Item, error) {
	return u.dao.GetItem(ctx, itemId)
}
