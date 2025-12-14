package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type UserGet interface {
	GetUser(ctx context.Context, id string) (*model.User, error)
}

type userGet struct {
	userDAO dao.UserDAO
}

func NewUserGet(userDAO dao.UserDAO) UserGet {
	return &userGet{userDAO: userDAO}
}

// GetUser : 指定されたIDのユーザーを取得
func (ug *userGet) GetUser(ctx context.Context, id string) (*model.User, error) {
	user, err := ug.userDAO.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fail:userDAO.GetUser:%w", err)
	}

	return user, nil
}
