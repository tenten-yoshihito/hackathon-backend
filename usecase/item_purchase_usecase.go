package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type ItemPurchase interface {
	PurchaseItem(ctx context.Context, itemID string, buyerID string) error
}

type itemPurchase struct {
	itemDAO dao.ItemDAO
}

func NewItemPurchase(itemDAO dao.ItemDAO) ItemPurchase {
	return &itemPurchase{itemDAO: itemDAO}
}

func (u *itemPurchase) PurchaseItem(ctx context.Context, itemID string, buyerID string) error {
	// 商品が存在し、販売中かチェック
	item, err := u.itemDAO.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	if item.Status != model.StatusOnSale {
		return fmt.Errorf("item is not available for purchase")
	}

	// 購入処理
	if err := u.itemDAO.PurchaseItem(ctx, itemID, buyerID); err != nil {
		return fmt.Errorf("failed to purchase item: %w", err)
	}

	return nil
}
