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

var ErrInvalidItemRequest = errors.New("invalid item request")

type ItemRegister interface {
	RegisterItem(ctx context.Context, uid string, req *model.ItemCreateRequest) (string, error)
}

type itemRegister struct {
	itemDAO dao.ItemDAO
}

func NewItemRegister(dao dao.ItemDAO) ItemRegister {
	return &itemRegister{itemDAO: dao}
}

func (us *itemRegister) RegisterItem(ctx context.Context, uid string, req *model.ItemCreateRequest) (string, error) {

	if !req.IsValid() {
		return "", ErrInvalidItemRequest
	}

	// 商品IDを生成
	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	newItemID := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	newItem := model.Item{
		ItemId:      newItemID,
		UserId:      uid,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		ImageURLs:   req.ImageURLs,
		CreatedAt:   t,
		UpdatedAt:   t,
	}

	err := us.itemDAO.ItemInsert(ctx, &newItem)
	if err != nil {
		return "", fmt.Errorf("fail:itemDAO.ItemInsert: %w", err)
	}

	return newItemID, nil
}
