package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type ItemUpdate interface {
	UpdateItem(ctx context.Context, req *model.ItemUpdateRequest) error
}

type itemUpdate struct {
	itemDAO dao.ItemDAO
}

func NewItemUpdate(itemDAO dao.ItemDAO) ItemUpdate {
	return &itemUpdate{itemDAO: itemDAO}
}

func (u *itemUpdate) UpdateItem(ctx context.Context, req *model.ItemUpdateRequest) error {
	if !req.IsValid() {
		return fmt.Errorf("invalid request")
	}

	err := u.itemDAO.UpdateItem(ctx, req.ItemID, req.UserID, req.Name, req.Price, req.Description)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}
