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
	register usecase.ItemRegister
	list     usecase.ItemList
}

func NewItemController(r usecase.ItemRegister, l usecase.ItemList) *ItemController {
	return &ItemController{register: r, list: l}
}

func (c *ItemController) HandleItemRegister(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

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

	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	ctx := r.Context()
	items, err := c.list.GetItems(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch items", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}
