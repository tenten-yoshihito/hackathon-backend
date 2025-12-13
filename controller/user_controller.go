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
	get      usecase.UserGet
	update   usecase.UserUpdate
}

func NewUserController(
	r usecase.UserRegister,
	s usecase.UserSearch,
	g usecase.UserGet,
	u usecase.UserUpdate,
) *UserController {
	return &UserController{
		register: r,
		search:   s,
		get:      g,
		update:   u,
	}
}

// ユーザー登録 (POST /register)
func (c *UserController) HandleProfileRegister(w http.ResponseWriter, r *http.Request) {
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

		if errors.Is(err, model.ErrInvalidRequest) {
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
	ctx := r.Context()
	users, err := c.search.Search(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch users", err)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

// ユーザー取得 (GET /users/{id})
func (c *UserController) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.PathValue("id")

	if userID == "" {
		respondError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	user, err := c.get.GetUser(ctx, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found", err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// ユーザー情報更新 (PUT /users/me)
func (c *UserController) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 認証されたユーザーIDを取得
	uid, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	var req model.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := c.update.UpdateUser(ctx, uid, &req); err != nil {
		if errors.Is(err, model.ErrInvalidRequest) {
			respondError(w, http.StatusBadRequest, "Invalid input data", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Internal server error", err)
		}
		return
	}

	log.Printf("successfully updated user: id=%s", uid)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Profile updated successfully"})
}
