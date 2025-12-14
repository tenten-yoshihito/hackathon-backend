package usecase

import (
	"context"
	"db/service"
	"fmt"
)

type DescriptionGenerate interface {
	GenerateFromImageURL(ctx context.Context, imageURL string) (string, error)
}

type descriptionGenerate struct {
	geminiService service.GeminiService
}


func NewDescriptionGenerate(geminiService service.GeminiService) DescriptionGenerate {
	return &descriptionGenerate{geminiService: geminiService}
}

// GenerateFromImageURL : service層のGeminiServiceを用いて商品説明を生成する 
func (u *descriptionGenerate) GenerateFromImageURL(ctx context.Context, imageURL string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("image URL is required")
	}

	description, err := u.geminiService.GenerateDescriptionFromImageURL(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate description: %w", err)
	}

	return description, nil
}
