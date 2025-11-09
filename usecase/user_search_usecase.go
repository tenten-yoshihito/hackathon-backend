package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

// UserSearchUsecase handles user search business logic

type UserSearch interface {
	Search(ctx context.Context) ([]model.User, error)
}

type userSearch struct {
	userDAO dao.UserDAO
}

func NewUserSearch(us dao.UserDAO) UserSearch {
	return &userSearch{userDAO: us}
}

func (us *userSearch) Search(ctx context.Context) ([]model.User, error) {

	users, err := us.userDAO.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail:usecase.Search: %w", err)
	}

	return users, nil
}
