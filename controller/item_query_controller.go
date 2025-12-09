package controller

import (
	"db/middleware"
	"db/usecase"
	"log"
	"net/http"
)

// ItemQueryController handles read-only item operations
type ItemQueryController struct {
	list        usecase.ItemList
	myItemsList usecase.MyItemsList
	get         usecase.ItemGet
}

// NewItemQueryController creates a new ItemQueryController
func NewItemQueryController(
	list usecase.ItemList,
	myItemsList usecase.MyItemsList,
	get usecase.ItemGet,
) *ItemQueryController {
	return &ItemQueryController{
		list:        list,
		myItemsList: myItemsList,
		get:         get,
	}
}

// HandleItemList retrieves all items (GET /items)
func (c *ItemQueryController) HandleItemList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	items, err := c.list.GetItems(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch items", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

// HandleItemDetail retrieves item details (GET /items/{id})
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

// HandleMyItems retrieves the list of items listed by the authenticated user (GET /items/my)
func (c *ItemQueryController) HandleMyItems(w http.ResponseWriter, r *http.Request) {
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
