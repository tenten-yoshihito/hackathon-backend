package usecase

import (
	"context"
	"db/dao"
	"db/model"
	"fmt"
)

type LikeUsecase interface {
	ToggleLike(ctx context.Context, userID, itemID string) error
	GetLikedItems(ctx context.Context, userID string) ([]model.ItemSimple, error)
	GetLikedItemIDs(ctx context.Context, userID string) ([]string, error)
}

type likeUsecase struct {
	likeDAO dao.LikeDAO
}

func NewLikeUsecase(likeDAO dao.LikeDAO) LikeUsecase {
	return &likeUsecase{likeDAO: likeDAO}
}

func (u *likeUsecase) ToggleLike(ctx context.Context, userID, itemID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if itemID == "" {
		return fmt.Errorf("item ID is required")
	}

	err := u.likeDAO.ToggleLike(ctx, userID, itemID)
	if err != nil {
		return fmt.Errorf("failed to toggle like: %w", err)
	}

	return nil
}

func (u *likeUsecase) GetLikedItems(ctx context.Context, userID string) ([]model.ItemSimple, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	items, err := u.likeDAO.GetLikedItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get liked items: %w", err)
	}

	return items, nil
}

func (u *likeUsecase) GetLikedItemIDs(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	itemIDs, err := u.likeDAO.GetLikedItemIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get liked item IDs: %w", err)
	}

	return itemIDs, nil
}
