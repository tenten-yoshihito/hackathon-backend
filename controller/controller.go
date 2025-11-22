package controller

// SearchUserController handles user search requests
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
	return &UserController{
		register: r,
		search:   s,
	}
}

func (c *UserController) HandleProfileRegister(w http.ResponseWriter, r *http.Request) {
	// Firebase認証ミドルウェアで設定されたユーザーIDをコンテキストから取得
	ctx := r.Context()
	uid, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		log.Printf("fail: get user ID from context, %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// リクエストボディからUserCreateRequestをデコード
	var req *model.UserCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("fail:json.NewDecoder, %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.register.Register(ctx, uid, req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidRequest) {
			log.Printf("fail: invalid request by client, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			log.Printf("fail: internal server error, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	log.Printf("successfully created user with transaction: id=%s", uid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	res := map[string]string{"id": uid}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("fail: json.NewEncoder, %v\n", err)
		return
	}
}

func (c *UserController) handleSearch(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	users, err := c.search.Search(ctx)
	if err != nil {
		log.Printf("fail:internal server error,%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("fail: json.NewEncoder, %v\n", err)
	}

}

func (c *UserController) HandleUser(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:
		c.handleSearch(w, r)

	// case http.MethodPost:
	// 	c.HandleProfileRegister(w, r)

	default:
		log.Printf("MethodNotAllowed:%v\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
