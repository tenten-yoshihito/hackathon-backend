package model

import "time"

type Item struct {
	ItemId      string    `json:"id"`
	UserId      string    `json:"user_id"`
	Name        string    `json:"name"`
	Price       int       `json:"price"`
	Description string    `json:"description,omitempty"`
	ImageURLs   []string  `json:"image_urls"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ItemCreateRequest struct {
	Name        string   `json:"name"`
	Price       int      `json:"price"`
	Description string   `json:"description,omitempty"`
	ImageURLs   []string `json:"image_urls"`
}

// IsValid バリデーション
func (req *ItemCreateRequest) IsValid() bool {
	return req.Name != "" && req.Price >= 0 && 10 > len(req.ImageURLs) && len(req.ImageURLs) > 0
}
