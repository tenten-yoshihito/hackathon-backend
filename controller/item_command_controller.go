package controller

import (
	"db/middleware"
	"db/model"
	"db/usecase"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// ItemCommandController handles item creation, update, and purchase operations
type ItemCommandController struct {
	register usecase.ItemRegister
	update   usecase.ItemUpdate
	purchase usecase.ItemPurchase
}

// NewItemCommandController creates a new ItemCommandController
func NewItemCommandController(
	register usecase.ItemRegister,
	update usecase.ItemUpdate,
	purchase usecase.ItemPurchase,
) *ItemCommandController {
	return &ItemCommandController{
		register: register,
		update:   update,
		purchase: purchase,
	}
}

// HandleItemRegister creates a new item (POST /items)
func (c *ItemCommandController) HandleItemRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	var req model.ItemCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	newItemID, err := c.register.RegisterItem(ctx, uid, &req)

	if err != nil {
		if errors.Is(err, usecase.ErrInvalidItemRequest) {
			respondError(w, http.StatusBadRequest, "Invalid request", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Internal server error", err)
		}
		return
	}

	log.Printf("successfully created item: id=%s", newItemID)
	respondJSON(w, http.StatusCreated, map[string]string{"id": newItemID})
}

// HandleItemPurchase purchases an item (POST /items/{id}/purchase)
func (c *ItemCommandController) HandleItemPurchase(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	itemID := r.PathValue("id")

	if err := c.purchase.PurchaseItem(ctx, itemID, uid); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to purchase item", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "purchase successful"})
}

// HandleItemUpdate updates an item (PUT /items/{id})
func (c *ItemCommandController) HandleItemUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	// Get item ID from URL parameter
	itemID := r.PathValue("id")
	if itemID == "" {
		respondError(w, http.StatusBadRequest, "item ID is required", nil)
		return
	}

	// Parse request body
	var req model.ItemUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set ItemID and UserID from context and URL
	req.ItemID = itemID
	req.UserID = userID

	// Execute update (validation is done in usecase)
	err = c.update.UpdateItem(ctx, &req)
	if err != nil {
		log.Printf("failed to update item: %v\n", err)
		if errors.Is(err, model.ErrNotAuthorized) {
			respondError(w, http.StatusForbidden, "Not authorized to update this item", err)
			return
		}
		if errors.Is(err, model.ErrCannotUpdateSoldItem) {
			respondError(w, http.StatusBadRequest, "Cannot update sold item", err)
			return
		}
		if errors.Is(err, model.ErrInvalidUpdateRequest) {
			respondError(w, http.StatusBadRequest, "Invalid request", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update item", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Item updated successfully"})
}
