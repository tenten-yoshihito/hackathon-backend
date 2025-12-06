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

// NewDescriptionGenerate creates a new DescriptionGenerate usecase
func NewDescriptionGenerate(geminiService service.GeminiService) DescriptionGenerate {
	return &descriptionGenerate{geminiService: geminiService}
}

// GenerateFromImageURL generates a product description from an image URL
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
