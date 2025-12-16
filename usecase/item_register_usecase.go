package usecase

import (
	"context"
	"crypto/rand"
	"db/dao"
	"db/model"
	"db/service"
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
	itemDAO       dao.ItemDAO
	geminiService service.GeminiService
}

func NewItemRegister(dao dao.ItemDAO, geminiService service.GeminiService) ItemRegister {
	return &itemRegister{itemDAO: dao, geminiService: geminiService}
}

func (us *itemRegister) RegisterItem(ctx context.Context, uid string, req *model.ItemCreateRequest) (string, error) {

	if !req.IsValid() {
		return "", ErrInvalidItemRequest
	}
	// 商品説明をベクトル化
	textToEmbed := fmt.Sprintf("%s\n%s", req.Name, req.Description)
	embedding, err := us.geminiService.GenerateEmbedding(ctx, textToEmbed)
	if err != nil {
		fmt.Printf("Warning: failed to generate embedding: %v\n", err)
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
		Embedding:   embedding,
		ImageURLs:   req.ImageURLs,
		CreatedAt:   t,
		UpdatedAt:   t,
	}
	err = us.itemDAO.ItemInsert(ctx, &newItem)
	if err != nil {
		return "", fmt.Errorf("fail:itemDAO.ItemInsert: %w", err)
	}

	return newItemID, nil
}
