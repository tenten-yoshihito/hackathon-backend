package usecase

import (
	"context"
	"crypto/rand"
	"db/cache"
	"db/dao"
	"db/model"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid"
)

type ItemPurchase interface {
	PurchaseItem(ctx context.Context, itemID string, buyerID string) error
}

type itemPurchase struct {
	itemDAO         dao.ItemDAO
	notificationDAO dao.NotificationDAO
	embeddingCache  *cache.EmbeddingCache
}

func NewItemPurchase(itemDAO dao.ItemDAO, notificationDAO dao.NotificationDAO, embeddingCache *cache.EmbeddingCache) ItemPurchase {
	return &itemPurchase{
		itemDAO:         itemDAO,
		notificationDAO: notificationDAO,
		embeddingCache:  embeddingCache,
	}
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

	// 出品者に通知を作成
	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	notificationID := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	notification := &model.Notification{
		Id:        notificationID,
		UserId:    item.UserId, // 出品者
		Type:      "purchase",
		ItemId:    itemID,
		ItemName:  item.Name,
		Message:   fmt.Sprintf("%sが購入されました", item.Name),
		IsRead:    false,
		CreatedAt: t,
	}

	// 通知作成が失敗しても購入は成功とする
	if err := u.notificationDAO.CreateNotification(ctx, notification); err != nil {
		log.Printf("Warning: failed to create notification: %v\n", err)
	}

	// キャッシュからベクトルを削除（おすすめから除外）
	u.embeddingCache.Delete(itemID)

	return nil
}
