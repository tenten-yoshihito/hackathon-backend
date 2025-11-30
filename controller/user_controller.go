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

type UserController struct {
	register usecase.UserRegister
	search   usecase.UserSearch
}

func NewUserController(r usecase.UserRegister, s usecase.UserSearch) *UserController {
	return &UserController{register: r, search: s}
}

// ユーザー登録 (POST /register)
func (c *UserController) HandleProfileRegister(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	ctx := r.Context()

	uid, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	var req model.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := c.register.Register(ctx, uid, &req); err != nil {

		if errors.Is(err, usecase.ErrInvalidRequest) {
			respondError(w, http.StatusBadRequest, "Invalid input data", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Internal server error", err)
		}
		return
	}

	log.Printf("successfully created user: id=%s", uid)
	respondJSON(w, http.StatusCreated, map[string]string{"id": uid})
}

// ユーザー検索 (GET /user)
func (c *UserController) HandleSearchUser(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	ctx := r.Context()
	users, err := c.search.Search(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch users", err)
		return
	}

	respondJSON(w, http.StatusOK, users)
}
