package cache

import (
	"context"
	"db/dao"
	"fmt"
	"log"
	"sync"
	"time"
)

// EmbeddingCache : ベクトルのインメモリキャッシュ
type EmbeddingCache struct {
	mu      sync.RWMutex
	data    map[string][]float32
	itemDAO dao.ItemDAO
}

// NewEmbeddingCache : キャッシュの初期化と自動ロード
func NewEmbeddingCache(itemDAO dao.ItemDAO) *EmbeddingCache {
	cache := &EmbeddingCache{
		data:    make(map[string][]float32),
		itemDAO: itemDAO,
	}

	// 起動時に一度ロード
	if err := cache.Reload(context.Background()); err != nil {
		log.Printf("Warning: failed to load embeddings on startup: %v", err)
	} else {
		log.Printf("Embedding cache initialized with %d items", len(cache.data))
	}

	return cache
}

// Reload : DBから全ベクトルを再ロード
func (c *EmbeddingCache) Reload(ctx context.Context) error {
	start := time.Now()

	embeddings, err := c.itemDAO.GetAllItemEmbeddings(ctx)
	if err != nil {
		return fmt.Errorf("failed to load embeddings: %w", err)
	}

	c.mu.Lock()
	c.data = embeddings
	c.mu.Unlock()

	log.Printf("Embedding cache reloaded: %d items in %v", len(embeddings), time.Since(start))
	return nil
}

// Get : 全ベクトルを取得（読み取り専用）
func (c *EmbeddingCache) Get() map[string][]float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// コピーして返す（外部での変更を防ぐ）
	result := make(map[string][]float32, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Set : 特定の商品のベクトルを更新
func (c *EmbeddingCache) Set(itemID string, embedding []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(embedding) > 0 {
		c.data[itemID] = embedding
		log.Printf("Cache updated for item: %s", itemID)
	}
}

// Delete : 特定の商品のベクトルを削除
func (c *EmbeddingCache) Delete(itemID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, itemID)
	log.Printf("Cache deleted for item: %s", itemID)
}

// GetCount : キャッシュされているアイテム数
func (c *EmbeddingCache) GetCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}
