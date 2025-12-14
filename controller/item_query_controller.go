package controller

import (
	"db/middleware"
	"db/model"
	"db/usecase"
	"log"
	"net/http"
)

// ItemQueryController : 商品情報を取得するコントローラ
type ItemQueryController struct {
	list          usecase.ItemList
	myItemsList   usecase.MyItemsList
	userItemsList usecase.UserItemsList
	get           usecase.ItemGet
}

func NewItemQueryController(
	list usecase.ItemList,
	myItemsList usecase.MyItemsList,
	userItemsList usecase.UserItemsList,
	get usecase.ItemGet,
) *ItemQueryController {
	return &ItemQueryController{
		list:          list,
		myItemsList:   myItemsList,
		userItemsList: userItemsList,
		get:           get,
	}
}

// HandleItemList : 商品一覧を取得する
func (c *ItemQueryController) HandleItemList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check for search keyword in query parameter
	keyword := r.URL.Query().Get("name")

	var items []model.ItemSimple
	var err error

	if keyword != "" {
		// Search items by keyword
		items, err = c.list.SearchItems(ctx, keyword)
	} else {
		// Get all items
		items, err = c.list.GetItems(ctx)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch items", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

// HandleItemDetail : 商品の詳細を取得する
func (c *ItemQueryController) HandleItemDetail(w http.ResponseWriter, r *http.Request) {
	itemID := r.PathValue("id")

	ctx := r.Context()
	item, err := c.get.GetItem(ctx, itemID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch item", err)
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// HandleMyItems : 自分の出品商品一覧を取得 (GET /items/my)
func (c *ItemQueryController) HandleMyItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by Firebase auth middleware)
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Printf("authentication failed: %v\n", err)
		respondError(w, http.StatusUnauthorized, "User not authenticated", err)
		return
	}

	items, err := c.myItemsList.GetMyItems(ctx, userID)
	if err != nil {
		log.Printf("failed to get my items: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to get my items", err)
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// HandleUserItems : 特定のユーザーの出品商品一覧を取得 (GET /users/{userId}/items)
func (c *ItemQueryController) HandleUserItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter
	userID := r.PathValue("userId")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	items, err := c.userItemsList.GetUserItems(ctx, userID)
	if err != nil {
		log.Printf("failed to get user items: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user items", err)
		return
	}

	respondJSON(w, http.StatusOK, items)
}
