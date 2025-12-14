package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/genai"
)

type GeminiService interface {
	GenerateDescriptionFromImageURL(ctx context.Context, imageURL string) (string, error)
}

type geminiService struct {
	apiKey string
}

func NewGeminiService(apiKey string) GeminiService {
	return &geminiService{
		apiKey: apiKey,
	}
}

func (s *geminiService) GenerateDescriptionFromImageURL(ctx context.Context, imageURL string) (string, error) {
	// Download image from URL
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Create Gemini client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  s.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Prepare prompt and image
	prompt := "あなたはフリマアプリの商品説明を書くアシスタントです。この画像の商品について、魅力的な説明文を200文字以内の日本語で書いてください。商品の特徴、状態、用途を含めてください。前置きや見出しは不要で、説明文のみを出力してください。"

	// Generate content with image
	resp2, err := client.Models.GenerateContent(ctx, "models/gemini-2.5-flash",
		[]*genai.Content{
			{
				Role: genai.RoleUser,
				Parts: []*genai.Part{
					genai.NewPartFromText(prompt),
					genai.NewPartFromBytes(imageData, "image/jpeg"),
				},
			},
		},
		nil, 
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract text from response
	if len(resp2.Candidates) == 0 || len(resp2.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Get the first text part
	text := ""
	for _, part := range resp2.Candidates[0].Content.Parts {
		if part.Text != "" {
			text += part.Text
		}
	}

	if text == "" {
		return "", fmt.Errorf("no text content in response")
	}

	return text, nil
}
