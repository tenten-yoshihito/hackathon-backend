package controller

import (
	"db/usecase"
	"net/http"
)

type RecommendController struct {
	recommendUsecase usecase.RecommendUsecase
}

func NewRecommendController(u usecase.RecommendUsecase) *RecommendController {
	return &RecommendController{recommendUsecase: u}
}

// HandleGetRecommendations : その商品に似た商品を提案 GET /items/{id}/recommend
func (c *RecommendController) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemID := r.PathValue("id")

	// 4件表示
	items, err := c.recommendUsecase.GetSimilarItems(ctx, itemID, 4)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get recommendations", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

// HandleGetPersonalizedRecommendations : ユーザーの好みに合わせたおすすめ GET /items/recommend
func (c *RecommendController) HandleGetPersonalizedRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// AuthMiddlewareでセットされたuserIDを取得
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// 10件表示
	items, err := c.recommendUsecase.GetPersonalizedRecommendations(ctx, userID, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get personalized recommendations", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}