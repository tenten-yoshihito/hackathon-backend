package usecase

import (
	"context"
	"db/cache"
	"db/model"
	"errors"
	"testing"
)

// MockNotificationDAO : dao.NotificationDAO のモック
type MockNotificationDAO struct {
	CreateNotificationFunc   func(ctx context.Context, notification *model.Notification) error
	GetUserNotificationsFunc func(ctx context.Context, userID string, limit int) ([]model.Notification, error)
	GetUnreadCountFunc       func(ctx context.Context, userID string) (int, error)
	MarkAsReadFunc           func(ctx context.Context, notificationID string, userID string) error
	MarkAllAsReadFunc        func(ctx context.Context, userID string) error
}

func (m *MockNotificationDAO) CreateNotification(ctx context.Context, notification *model.Notification) error {
	if m.CreateNotificationFunc != nil {
		return m.CreateNotificationFunc(ctx, notification)
	}
	return nil
}

func (m *MockNotificationDAO) GetUserNotifications(ctx context.Context, userID string, limit int) ([]model.Notification, error) {
	if m.GetUserNotificationsFunc != nil {
		return m.GetUserNotificationsFunc(ctx, userID, limit)
	}
	return nil, nil
}

func (m *MockNotificationDAO) MarkAsRead(ctx context.Context, notificationID string, userID string) error {
	if m.MarkAsReadFunc != nil {
		return m.MarkAsReadFunc(ctx, notificationID, userID)
	}
	return nil
}

func (m *MockNotificationDAO) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	if m.GetUnreadCountFunc != nil {
		return m.GetUnreadCountFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockNotificationDAO) MarkAllAsRead(ctx context.Context, userID string) error {
	if m.MarkAllAsReadFunc != nil {
		return m.MarkAllAsReadFunc(ctx, userID)
	}
	return nil
}

func TestItemPurchase_PurchaseItem(t *testing.T) {
	// Setup
	validItem := &model.Item{
		ItemId: "item1",
		UserId: "seller1",
		Status: model.StatusOnSale,
		Price:  1000,
	}
	soldItem := &model.Item{
		ItemId: "item2",
		UserId: "seller1",
		Status: model.StatusSold,
		Price:  1000,
	}

	tests := []struct {
		name                 string
		itemID               string
		buyerID              string
		mockItemDAO          *MockItemDAO
		mockNotificationDAO  *MockNotificationDAO
		wantErr              bool
		wantEmbeddingDeleted bool
	}{
		{
			name:    "成功: 購入成功",
			itemID:  "item1",
			buyerID: "buyer1",
			mockItemDAO: &MockItemDAO{
				GetItemFunc: func(ctx context.Context, itemID string) (*model.Item, error) {
					return validItem, nil
				},
				PurchaseItemFunc: func(ctx context.Context, itemID string, buyerID string) error {
					return nil
				},
				GetAllItemEmbeddingsFunc: func(ctx context.Context) (map[string][]float32, error) {
					return map[string][]float32{"item1": {0.1, 0.2}}, nil
				},
			},
			mockNotificationDAO: &MockNotificationDAO{
				CreateNotificationFunc: func(ctx context.Context, notification *model.Notification) error {
					return nil
				},
			},
			wantErr:              false,
			wantEmbeddingDeleted: true,
		},
		{
			name:    "失敗: 商品が存在しない",
			itemID:  "invalid",
			buyerID: "buyer1",
			mockItemDAO: &MockItemDAO{
				GetItemFunc: func(ctx context.Context, itemID string) (*model.Item, error) {
					return nil, errors.New("not found")
				},
				GetAllItemEmbeddingsFunc: func(ctx context.Context) (map[string][]float32, error) {
					return map[string][]float32{}, nil
				},
			},
			mockNotificationDAO: &MockNotificationDAO{},
			wantErr:             true,
		},
		{
			name:    "失敗: 売り切れ",
			itemID:  "item2",
			buyerID: "buyer1",
			mockItemDAO: &MockItemDAO{
				GetItemFunc: func(ctx context.Context, itemID string) (*model.Item, error) {
					return soldItem, nil
				},
				GetAllItemEmbeddingsFunc: func(ctx context.Context) (map[string][]float32, error) {
					return map[string][]float32{}, nil
				},
			},
			mockNotificationDAO: &MockNotificationDAO{},
			wantErr:             true,
		},
		{
			name:    "失敗: DAOエラー",
			itemID:  "item1",
			buyerID: "buyer1",
			mockItemDAO: &MockItemDAO{
				GetItemFunc: func(ctx context.Context, itemID string) (*model.Item, error) {
					return validItem, nil
				},
				PurchaseItemFunc: func(ctx context.Context, itemID string, buyerID string) error {
					return errors.New("db error")
				},
				GetAllItemEmbeddingsFunc: func(ctx context.Context) (map[string][]float32, error) {
					return map[string][]float32{"item1": {0.1, 0.2}}, nil
				},
			},
			mockNotificationDAO: &MockNotificationDAO{},
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// EmbeddingCache の初期化
			// MockItemDAOを使って初期ロードが行われる
			embeddingCache := cache.NewEmbeddingCache(tt.mockItemDAO)

			// キャッシュの状態セット (成功ケースで削除確認するため)
			if tt.wantEmbeddingDeleted {
				embeddingCache.Set(tt.itemID, []float32{0.1, 0.2})
			}

			u := NewItemPurchase(tt.mockItemDAO, tt.mockNotificationDAO, embeddingCache)
			err := u.PurchaseItem(context.Background(), tt.itemID, tt.buyerID)

			if (err != nil) != tt.wantErr {
				t.Errorf("PurchaseItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			// キャッシュから削除されたか確認
			if tt.wantEmbeddingDeleted && err == nil {
				embeddings := embeddingCache.Get()
				if _, ok := embeddings[tt.itemID]; ok {
					t.Errorf("Embedding should be deleted from cache, but it exists")
				}
			}
		})
	}
}
