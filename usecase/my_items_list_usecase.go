package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type MyItemsList interface {
	GetMyItems(ctx context.Context, userID string) ([]model.ItemSimple, error)
}

type myItemsList struct {
	itemDAO dao.ItemDAO
}

func NewMyItemsList(itemDAO dao.ItemDAO) MyItemsList {
	return &myItemsList{itemDAO: itemDAO}
}

func (u *myItemsList) GetMyItems(ctx context.Context, userID string) ([]model.ItemSimple, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	items, err := u.itemDAO.GetMyItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get my items: %w", err)
	}

	return items, nil
}
