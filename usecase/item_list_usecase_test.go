package usecase

import (
	"context"
	"db/model"
	"errors"
	"testing"
)

// MockItemDAO : dao.ItemDAO のモック
type MockItemDAO struct {
	// 各メソッドの戻り値を制御するための関数フィールドなどを必要に応じて追加
	GetItemListFunc func(ctx context.Context, limit int, offset int) ([]model.ItemSimple, error)
	SearchItemsFunc func(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error)

	// 未使用メソッドのスタブ (コンパイルエラー回避のため)
	ItemInsertFunc           func(ctx context.Context, item *model.Item) error
	GetMyItemsFunc           func(ctx context.Context, sellerID string) ([]model.ItemSimple, error)
	GetUserItemsFunc         func(ctx context.Context, userID string) ([]model.ItemSimple, error)
	GetItemFunc              func(ctx context.Context, itemID string) (*model.Item, error)
	GetItemsByIDsFunc        func(ctx context.Context, itemIDs []string) ([]model.ItemSimple, error)
	PurchaseItemFunc         func(ctx context.Context, itemID string, buyerID string) error
	UpdateItemFunc           func(ctx context.Context, itemID string, userID string, name string, price int, description string, imageURLs []string, embedding []float32) error
	GetAllItemEmbeddingsFunc func(ctx context.Context) (map[string][]float32, error)
	GetItemEmbeddingFunc     func(ctx context.Context, itemID string) ([]float32, error)
}

func (m *MockItemDAO) ItemInsert(ctx context.Context, item *model.Item) error {
	if m.ItemInsertFunc != nil {
		return m.ItemInsertFunc(ctx, item)
	}
	return nil
}

func (m *MockItemDAO) GetItemList(ctx context.Context, limit int, offset int) ([]model.ItemSimple, error) {
	if m.GetItemListFunc != nil {
		return m.GetItemListFunc(ctx, limit, offset)
	}
	return nil, nil
}

func (m *MockItemDAO) SearchItems(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error) {
	if m.SearchItemsFunc != nil {
		return m.SearchItemsFunc(ctx, keyword, limit, offset)
	}
	return nil, nil
}

func (m *MockItemDAO) GetMyItems(ctx context.Context, sellerID string) ([]model.ItemSimple, error) {
	if m.GetMyItemsFunc != nil {
		return m.GetMyItemsFunc(ctx, sellerID)
	}
	return nil, nil
}

func (m *MockItemDAO) GetUserItems(ctx context.Context, userID string) ([]model.ItemSimple, error) {
	if m.GetUserItemsFunc != nil {
		return m.GetUserItemsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockItemDAO) GetItem(ctx context.Context, itemID string) (*model.Item, error) {
	if m.GetItemFunc != nil {
		return m.GetItemFunc(ctx, itemID)
	}
	return nil, nil
}

func (m *MockItemDAO) GetItemsByIDs(ctx context.Context, itemIDs []string) ([]model.ItemSimple, error) {
	if m.GetItemsByIDsFunc != nil {
		return m.GetItemsByIDsFunc(ctx, itemIDs)
	}
	return nil, nil
}

func (m *MockItemDAO) PurchaseItem(ctx context.Context, itemID string, buyerID string) error {
	if m.PurchaseItemFunc != nil {
		return m.PurchaseItemFunc(ctx, itemID, buyerID)
	}
	return nil
}

func (m *MockItemDAO) UpdateItem(ctx context.Context, itemID string, userID string, name string, price int, description string, imageURLs []string, embedding []float32) error {
	if m.UpdateItemFunc != nil {
		return m.UpdateItemFunc(ctx, itemID, userID, name, price, description, imageURLs, embedding)
	}
	return nil
}

func (m *MockItemDAO) GetAllItemEmbeddings(ctx context.Context) (map[string][]float32, error) {
	if m.GetAllItemEmbeddingsFunc != nil {
		return m.GetAllItemEmbeddingsFunc(ctx)
	}
	return nil, nil
}

func (m *MockItemDAO) GetItemEmbedding(ctx context.Context, itemID string) ([]float32, error) {
	if m.GetItemEmbeddingFunc != nil {
		return m.GetItemEmbeddingFunc(ctx, itemID)
	}
	return nil, nil
}

func TestItemList_GetItems(t *testing.T) {
	mockItems := []model.ItemSimple{
		{ItemId: "1", Name: "Item 1", Price: 100},
		{ItemId: "2", Name: "Item 2", Price: 200},
	}

	tests := []struct {
		name    string
		mockDAO *MockItemDAO
		limit   int
		offset  int
		want    []model.ItemSimple
		wantErr bool
	}{
		{
			name: "成功: 一覧取得",
			mockDAO: &MockItemDAO{
				GetItemListFunc: func(ctx context.Context, limit int, offset int) ([]model.ItemSimple, error) {
					return mockItems, nil
				},
			},
			limit:   10,
			offset:  0,
			want:    mockItems,
			wantErr: false,
		},
		{
			name: "失敗: DAOエラー",
			mockDAO: &MockItemDAO{
				GetItemListFunc: func(ctx context.Context, limit int, offset int) ([]model.ItemSimple, error) {
					return nil, errors.New("db error")
				},
			},
			limit:   10,
			offset:  0,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewItemList(tt.mockDAO)
			got, err := u.GetItems(context.Background(), tt.limit, tt.offset)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != len(tt.want) {
				t.Errorf("GetItems() got length = %v, want length %v", len(got), len(tt.want))
			}
		})
	}
}

func TestItemList_SearchItems(t *testing.T) {
	mockItems := []model.ItemSimple{
		{ItemId: "1", Name: "Search Result 1", Price: 100},
	}

	tests := []struct {
		name    string
		mockDAO *MockItemDAO
		keyword string
		limit   int
		offset  int
		want    []model.ItemSimple
		wantErr bool
	}{
		{
			name: "成功: 検索成功",
			mockDAO: &MockItemDAO{
				SearchItemsFunc: func(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error) {
					if keyword == "test" {
						return mockItems, nil
					}
					return []model.ItemSimple{}, nil
				},
			},
			keyword: "test",
			limit:   10,
			offset:  0,
			want:    mockItems,
			wantErr: false,
		},
		{
			name: "失敗: キーワード空",
			mockDAO: &MockItemDAO{
				SearchItemsFunc: func(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error) {
					return nil, nil // 呼ばれないはず
				},
			},
			keyword: "  ", // 空白のみ
			limit:   10,
			offset:  0,
			want:    nil,
			wantErr: true, // バリデーションエラー
		},
		{
			name: "失敗: DAOエラー",
			mockDAO: &MockItemDAO{
				SearchItemsFunc: func(ctx context.Context, keyword string, limit int, offset int) ([]model.ItemSimple, error) {
					return nil, errors.New("search error")
				},
			},
			keyword: "error",
			limit:   10,
			offset:  0,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewItemList(tt.mockDAO)
			got, err := u.SearchItems(context.Background(), tt.keyword, tt.limit, tt.offset)

			if (err != nil) != tt.wantErr {
				t.Errorf("SearchItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != len(tt.want) {
				t.Errorf("SearchItems() got length = %v, want length %v", len(got), len(tt.want))
			}
		})
	}
}
