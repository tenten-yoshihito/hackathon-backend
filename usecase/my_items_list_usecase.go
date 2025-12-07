package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type MyItemsList interface {
	Execute(ctx context.Context, userID string) ([]model.ItemSimple, error)
}

type myItemsList struct {
	itemDAO dao.ItemDAO
}

// NewMyItemsList creates a new MyItemsList usecase
func NewMyItemsList(itemDAO dao.ItemDAO) MyItemsList {
	return &myItemsList{itemDAO: itemDAO}
}

// Execute retrieves the list of items listed by the user
func (u *myItemsList) Execute(ctx context.Context, userID string) ([]model.ItemSimple, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	items, err := u.itemDAO.GetMyItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get my items: %w", err)
	}

	return items, nil
}
