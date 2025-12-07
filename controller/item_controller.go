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
	descriptionGenerate usecase.DescriptionGenerate
}

// ItemControllerConfig holds dependencies for ItemController
type ItemControllerConfig struct {
	Register            usecase.ItemRegister
	List                usecase.ItemList
	MyItemsList         usecase.MyItemsList
	Get                 usecase.ItemGet
	Purchase            usecase.ItemPurchase
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
		descriptionGenerate: config.DescriptionGenerate,
	}
}

func (c *ItemController) HandleItemRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Printf("fail: GetUserIDFromContext, %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req model.ItemCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("fail: json.NewDecoder, %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newItemID, err := c.register.RegisterItem(ctx, uid, &req)

	if err != nil {
		if errors.Is(err, usecase.ErrInvalidItemRequest) {
			log.Printf("fail: invalid request, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			log.Printf("fail: internal server error, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	log.Printf("successfully created item: id=%s", newItemID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	res := map[string]string{"id": newItemID}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("fail: json.NewEncoder, %v\n", err)
	}
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
		log.Printf("fail: GetUserIDFromContext, %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	itemID := r.PathValue("id")

	if err := c.purchase.PurchaseItem(ctx, itemID, uid); err != nil {
		log.Printf("fail: purchase item, %v\n", err)
		respondError(w, http.StatusBadRequest, err.Error(), err)
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
