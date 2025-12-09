package controller

import (
	"db/usecase"
	"encoding/json"
	"log"
	"net/http"
)

// ItemAIController handles AI-powered item operations
type ItemAIController struct {
	descriptionGenerate usecase.DescriptionGenerate
}

// NewItemAIController creates a new ItemAIController
func NewItemAIController(descriptionGenerate usecase.DescriptionGenerate) *ItemAIController {
	return &ItemAIController{
		descriptionGenerate: descriptionGenerate,
	}
}

// HandleGenerateDescription generates item description using AI (POST /items/generate-description)
func (c *ItemAIController) HandleGenerateDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		ImageURL string `json:"image_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.ImageURL == "" {
		respondError(w, http.StatusBadRequest, "image_url is required", nil)
		return
	}

	description, err := c.descriptionGenerate.GenerateFromImageURL(ctx, req.ImageURL)
	if err != nil {
		log.Printf("fail: generate description, %v\n", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate description", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"description": description})
}
