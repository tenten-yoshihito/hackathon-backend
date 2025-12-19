package usecase

import (
	"context"
	"db/cache"
	"db/dao"
	"db/model"
	"db/service"
	"fmt"
)

type ItemUpdate interface {
	UpdateItem(ctx context.Context, req *model.ItemUpdateRequest) error
}

type itemUpdate struct {
	itemDAO        dao.ItemDAO
	geminiService  service.GeminiService
	embeddingCache *cache.EmbeddingCache
}

func NewItemUpdate(itemDAO dao.ItemDAO, geminiService service.GeminiService, embeddingCache *cache.EmbeddingCache) ItemUpdate {
	return &itemUpdate{itemDAO: itemDAO, geminiService: geminiService, embeddingCache: embeddingCache}
}

func (u *itemUpdate) UpdateItem(ctx context.Context, req *model.ItemUpdateRequest) error {
	if !req.IsValid() {
		return fmt.Errorf("invalid request")
	}
	// 商品説明をベクトル化
	textToEmbed := fmt.Sprintf("%s\n%s", req.Name, req.Description)
	embedding, err := u.geminiService.GenerateEmbedding(ctx, textToEmbed)
	if err != nil {
		fmt.Printf("Warning: failed to update embedding: %v\n", err)
	}
	err = u.itemDAO.UpdateItem(ctx, req.ItemID, req.UserID, req.Name, req.Price, req.Description, req.ImageURLs, embedding)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	// キャッシュも即時更新
	u.embeddingCache.Set(req.ItemID, embedding)

	return nil
}
