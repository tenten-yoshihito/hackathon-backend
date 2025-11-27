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
	ctx := r.Context()

	uid, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	var req model.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := c.register.Register(ctx, uid, &req); err != nil {

		if errors.Is(err, usecase.ErrInvalidRequest) {
			c.respondError(w, http.StatusBadRequest, "Invalid input data", err)
		} else {
			c.respondError(w, http.StatusInternalServerError, "Internal server error", err)
		}
		return
	}

	log.Printf("successfully created user: id=%s", uid)
	c.respondJSON(w, http.StatusCreated, map[string]string{"id": uid})
}

// ユーザー検索 (GET /user)
func (c *UserController) HandleSearchUser(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		c.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	ctx := r.Context()
	users, err := c.search.Search(ctx)
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, "Failed to fetch users", err)
		return
	}

	c.respondJSON(w, http.StatusOK, users)
}

// ---------------------------------------------------------
//  共通ヘルパー関数 
// ---------------------------------------------------------

// respondJSON : 正常系レスポンスを返す
func (c *UserController) respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("fail: response encoding, %v\n", err)
	}
}

// respondError : エラーレスポンスを返す (ログ出力)
func (c *UserController) respondError(w http.ResponseWriter, status int, message string, err error) {
	log.Printf("error: %s: %v", message, err) 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
