package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"errors"
	"fmt"
)

// UserRegisterUsecase handles user registration business logic

var ErrInvalidRequest = errors.New("invalid request")

type UserRegister interface {
	Register(ctx context.Context, uid string, req *model.UserCreateRequest) error
}

type userRegister struct {
	userDAO dao.UserDAO
}

// NewUserRegister :daoからUserDAOを受けとり、userRegister構造体を生成して返す (この構造体にRegisterメソッドがある)
func NewUserRegister(us dao.UserDAO) UserRegister {
	return &userRegister{userDAO: us}
}

func (us *userRegister) Register(ctx context.Context, uid string, req *model.UserCreateRequest) error {

	if !req.IsValid() {
		return ErrInvalidRequest
	}

	newUser := model.User{
		Id:      uid,
		Name:    req.Name,
		Age:     req.Age,
		Email:   req.Email,
		IconURL: req.IconURL,
	}
	err := us.userDAO.DBInsert(ctx, &newUser)
	if err != nil {
		return fmt.Errorf("fail:us.userDAO.DBInsert:%w", err)
	}

	return nil
}
