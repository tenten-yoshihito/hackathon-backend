package usecase

import (
	"context"
	"db/cache"
	"db/dao"
	"db/model"
	"fmt"
	"math"
	"sort"
)

type RecommendUsecase interface {
	GetSimilarItems(ctx context.Context, targetItemID string, limit int) ([]model.ItemSimple, error)
	GetPersonalizedRecommendations(ctx context.Context, userID string, limit int) ([]model.ItemSimple, error)
}

type recommendUsecase struct {
	itemDAO        dao.ItemDAO
	likeDAO        dao.LikeDAO
	embeddingCache *cache.EmbeddingCache
}

func NewRecommendUsecase(itemDAO dao.ItemDAO, likeDAO dao.LikeDAO, embeddingCache *cache.EmbeddingCache) RecommendUsecase {
	return &recommendUsecase{
		itemDAO:        itemDAO,
		likeDAO:        likeDAO,
		embeddingCache: embeddingCache,
	}
}

// GetSimilarItems : 指定した商品に似ている商品を返す (Item-to-Item)
func (us *recommendUsecase) GetSimilarItems(ctx context.Context, targetItemID string, limit int) ([]model.ItemSimple, error) {
	// キャッシュから取得（超高速）
	allEmbeddings := us.embeddingCache.Get()

	targetVector, ok := allEmbeddings[targetItemID]
	if !ok {
		// キャッシュにない場合（SOLD商品など）はDBから直接取得
		vec, err := us.itemDAO.GetItemEmbedding(ctx, targetItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get target item embedding: %w", err)
		}
		if vec == nil {
			// ベクトル自体が存在しない（商品がない、またはベクトル未生成）
			return []model.ItemSimple{}, nil
		}
		targetVector = vec
	}

	// 類似度計算
	recommendations, err := us.calculateRanking(ctx, targetVector, allEmbeddings, limit, []string{targetItemID})
	if err != nil {
		return nil, err
	}

	return recommendations, nil
}

// GetPersonalizedRecommendations : ユーザーのいいね履歴からおすすめを返す (User-to-Item)
func (us *recommendUsecase) GetPersonalizedRecommendations(ctx context.Context, userID string, limit int) ([]model.ItemSimple, error) {
	// 1. いいねした商品IDを取得
	likedItemIDs, err := us.likeDAO.GetLikedItemIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get liked items: %w", err)
	}
	if len(likedItemIDs) == 0 {
		return []model.ItemSimple{}, nil
	}

	// 2. ベクトル取得（キャッシュから超高速）
	allEmbeddings := us.embeddingCache.Get()

	// 3. ユーザーベクトル（好みの平均）を作成
	var userVector []float32
	var count int

	for _, likedID := range likedItemIDs {
		if vec, ok := allEmbeddings[likedID]; ok && len(vec) > 0 {
			if userVector == nil {
				userVector = make([]float32, len(vec))
			}
			for i, v := range vec {
				userVector[i] += v
			}
			count++
		}
	}

	if count == 0 {
		return []model.ItemSimple{}, nil
	}

	// 平均化
	for i := range userVector {
		userVector[i] /= float32(count)
	}

	// 4. ランキング計算 (いいね済みの商品は除外)
	recommendations, err := us.calculateRanking(ctx, userVector, allEmbeddings, limit, likedItemIDs)
	if err != nil {
		return nil, err
	}

	return recommendations, nil
}

// 共通ロジック: ランキング計算と商品情報取得
func (us *recommendUsecase) calculateRanking(ctx context.Context, targetVec []float32, allEmbeddings map[string][]float32, limit int, excludeIDs []string) ([]model.ItemSimple, error) {
	type itemScore struct {
		ID    string
		Score float64
	}
	var scores []itemScore

	excludeMap := make(map[string]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	for id, vec := range allEmbeddings {
		if excludeMap[id] {
			continue
		}
		score := cosineSimilarity(targetVec, vec)
		scores = append(scores, itemScore{ID: id, Score: score})
	}

	// スコア順にソート
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Top N 抽出
	topN := limit
	if len(scores) < topN {
		topN = len(scores)
	}

	// Top NのIDを抽出
	topIDs := make([]string, topN)
	for i := 0; i < topN; i++ {
		topIDs[i] = scores[i].ID
	}

	// バルク取得（1回のクエリで全取得）
	results, err := us.itemDAO.GetItemsByIDs(ctx, topIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	return results, nil
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}
	var dot, normA, normB float64
	for i := range a {
		valA, valB := float64(a[i]), float64(b[i])
		dot += valA * valB
		normA += valA * valA
		normB += valB * valB
	}
	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
