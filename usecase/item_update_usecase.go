package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type ItemUpdate interface {
	Execute(ctx context.Context, req *model.ItemUpdateRequest) error
}

type itemUpdate struct {
	itemDAO dao.ItemDAO
}

// NewItemUpdate creates a new ItemUpdate usecase
func NewItemUpdate(itemDAO dao.ItemDAO) ItemUpdate {
	return &itemUpdate{itemDAO: itemDAO}
}

// Execute updates an item
func (u *itemUpdate) Execute(ctx context.Context, req *model.ItemUpdateRequest) error {
	if !req.IsValid() {
		return model.ErrInvalidUpdateRequest
	}

	err := u.itemDAO.UpdateItem(ctx, req.ItemID, req.UserID, req.Name, req.Price, req.Description)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}
