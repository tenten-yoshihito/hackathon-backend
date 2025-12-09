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

type ItemController struct {
	register            usecase.ItemRegister
	list                usecase.ItemList
	myItemsList         usecase.MyItemsList
	get                 usecase.ItemGet
	purchase            usecase.ItemPurchase
	update              usecase.ItemUpdate
	descriptionGenerate usecase.DescriptionGenerate
}

// ItemControllerConfig holds dependencies for ItemController
type ItemControllerConfig struct {
	Register            usecase.ItemRegister
	List                usecase.ItemList
	MyItemsList         usecase.MyItemsList
	Get                 usecase.ItemGet
	Purchase            usecase.ItemPurchase
	Update              usecase.ItemUpdate
	DescriptionGenerate usecase.DescriptionGenerate
}

// NewItemController creates a new ItemController with the given configuration
func NewItemController(config ItemControllerConfig) *ItemController {
	return &ItemController{
		register:            config.Register,
		list:                config.List,
		myItemsList:         config.MyItemsList,
		get:                 config.Get,
		purchase:            config.Purchase,
		update:              config.Update,
		descriptionGenerate: config.DescriptionGenerate,
	}
}

func (c *ItemController) HandleItemRegister(w http.ResponseWriter, r *http.Request) {
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

func (c *ItemController) HandleItemList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	items, err := c.list.GetItems(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch items", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

// HandleItemDetail : 商品詳細取得 (GET /items/{id})
func (c *ItemController) HandleItemDetail(w http.ResponseWriter, r *http.Request) {

	itemID := r.PathValue("id")

	ctx := r.Context()
	item, err := c.get.GetItem(ctx, itemID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch item", err)
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// HandleMyItems retrieves the list of items listed by the authenticated user (GET /items/my)
func (c *ItemController) HandleMyItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by Firebase auth middleware)
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Printf("authentication failed: %v\n", err)
		respondError(w, http.StatusUnauthorized, "User not authenticated", err)
		return
	}

	items, err := c.myItemsList.Execute(ctx, userID)
	if err != nil {
		log.Printf("failed to get my items: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to get my items", err)
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// HandleItemPurchase : 商品購入 (POST /items/{id}/purchase)
func (c *ItemController) HandleItemPurchase(w http.ResponseWriter, r *http.Request) {
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

// HandleGenerateDescription : AI商品説明生成 (POST /items/generate-description)
func (c *ItemController) HandleGenerateDescription(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var req struct {
		ImageURL string `json:"image_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.ImageURL == "" {
		respondError(w, http.StatusBadRequest, "image_url is required", nil)
		return
	}

	description, err := c.descriptionGenerate.GenerateFromImageURL(ctx, req.ImageURL)
	if err != nil {
		log.Printf("fail: generate description, %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate description", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"description": description})
}

// HandleItemUpdate handles PUT /items/:id
func (c *ItemController) HandleItemUpdate(w http.ResponseWriter, r *http.Request) {
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
	err = c.update.Execute(ctx, &req)
	if err != nil {
		log.Printf("failed to update item: %v\n", err)
		if errors.Is(err, errors.New("not authorized to update this item")) {
			respondError(w, http.StatusForbidden, "Not authorized to update this item", err)
			return
		}
		if errors.Is(err, errors.New("cannot update sold item")) {
			respondError(w, http.StatusBadRequest, "Cannot update sold item", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update item", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Item updated successfully"})
}
