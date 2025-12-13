package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type UserUpdate interface {
	UpdateUser(ctx context.Context, id string, req *model.UserUpdateRequest) error
}

type userUpdate struct {
	userDAO dao.UserDAO
}

// NewUserUpdate : UserUpdateの生成
func NewUserUpdate(userDAO dao.UserDAO) UserUpdate {
	return &userUpdate{userDAO: userDAO}
}

// UpdateUser : ユーザー情報を更新
func (uu *userUpdate) UpdateUser(ctx context.Context, id string, req *model.UserUpdateRequest) error {
	if !req.IsValid() {
		return model.ErrInvalidRequest
	}

	// 更新用のUserオブジェクトを作成
	user := &model.User{
		Id:      id,
		Name:    req.Name,
		Age:     req.Age,
		Bio:     req.Bio,
		IconURL: req.IconURL,
	}

	err := uu.userDAO.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("fail:userDAO.UpdateUser:%w", err)
	}

	return nil
}
