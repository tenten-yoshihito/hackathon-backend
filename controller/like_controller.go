package controller

import (
	"db/middleware"
	"db/usecase"
	"log"
	"net/http"
)

type LikeController struct {
	likeUsecase usecase.LikeUsecase
}

func NewLikeController(likeUsecase usecase.LikeUsecase) *LikeController {
	return &LikeController{
		likeUsecase: likeUsecase,
	}
}

// HandleToggleLike toggles a like on an item (POST /items/{id}/like)
func (c *LikeController) HandleToggleLike(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by Firebase auth middleware)
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Printf("authentication failed: %v\n", err)
		respondError(w, http.StatusUnauthorized, "User not authenticated", err)
		return
	}

	// Get item ID from URL parameter
	itemID := r.PathValue("id")
	if itemID == "" {
		respondError(w, http.StatusBadRequest, "Item ID is required", nil)
		return
	}

	err = c.likeUsecase.ToggleLike(ctx, userID, itemID)
	if err != nil {
		log.Printf("failed to toggle like: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to toggle like", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Like toggled successfully"})
}

// HandleGetLikedItems retrieves items that the user has liked (GET /items/liked)
func (c *LikeController) HandleGetLikedItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by Firebase auth middleware)
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Printf("authentication failed: %v\n", err)
		respondError(w, http.StatusUnauthorized, "User not authenticated", err)
		return
	}

	items, err := c.likeUsecase.GetLikedItems(ctx, userID)
	if err != nil {
		log.Printf("failed to get liked items: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to get liked items", err)
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// HandleGetLikedItemIDs retrieves the IDs of items that the user has liked (GET /items/liked-ids)
func (c *LikeController) HandleGetLikedItemIDs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by Firebase auth middleware)
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Printf("authentication failed: %v\n", err)
		respondError(w, http.StatusUnauthorized, "User not authenticated", err)
		return
	}

	itemIDs, err := c.likeUsecase.GetLikedItemIDs(ctx, userID)
	if err != nil {
		log.Printf("failed to get liked item IDs: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to get liked item IDs", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"liked_item_ids": itemIDs})
}
