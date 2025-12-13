package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type UserItemsList interface {
	GetUserItems(ctx context.Context, userID string) ([]model.ItemSimple, error)
}

type userItemsList struct {
	itemDAO dao.ItemDAO
}

// NewUserItemsList creates a new UserItemsList usecase
func NewUserItemsList(itemDAO dao.ItemDAO) UserItemsList {
	return &userItemsList{itemDAO: itemDAO}
}

// GetUserItems retrieves the list of items listed by a specific user
func (u *userItemsList) GetUserItems(ctx context.Context, userID string) ([]model.ItemSimple, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	items, err := u.itemDAO.GetUserItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user items: %w", err)
	}

	return items, nil
}
