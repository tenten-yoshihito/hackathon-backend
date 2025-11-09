package usecase

import (
	"context"
	"crypto/rand"
	"db/dao"
	"db/model"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"
)

// UserRegisterUsecase handles user registration business logic

var ErrInvalidRequest = errors.New("invalid request")

type UserRegister interface {
	Register(ctx context.Context, req *model.UserCreateRequest) (string, error)
}

type userRegister struct {
	userDAO dao.UserDAO
}

func NewUserRegister(us dao.UserDAO) UserRegister {
	return &userRegister{userDAO: us}
}

func (us *userRegister) Register(ctx context.Context, req *model.UserCreateRequest) (string, error) {

	if !req.IsValid() {
		return "", ErrInvalidRequest
	}

	entropy := ulid.Monotonic(rand.Reader, 0)
	newID := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	newUser := model.User{
		Id:   newID.String(),
		Name: req.Name,
		Age:  req.Age,
	}
	err := us.userDAO.DBInsert(ctx, &newUser)
	if err != nil {
		return "", fmt.Errorf("fail:us.userDAO.DBInsert:%w", err)
	}

	return newUser.Id, nil
}
