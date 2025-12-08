package model

import "time"

// Status constants
const (
	StatusOnSale = "ON_SALE"
	StatusSold   = "SOLD"
)

type Item struct {
	ItemId      string    `json:"id"`
	UserId      string    `json:"user_id"`
	Name        string    `json:"name"`
	Price       int       `json:"price"`
	Description string    `json:"description,omitempty"`
	ImageURLs   []string  `json:"image_urls"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ItemCreateRequest struct {
	Name        string   `json:"name"`
	Price       int      `json:"price"`
	Description string   `json:"description,omitempty"`
	ImageURLs   []string `json:"image_urls"`
}

type ItemUpdateRequest struct {
	ItemID      string `json:"item_id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description,omitempty"`
}

type ItemSimple struct {
	ItemId   string `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	ImageURL string `json:"image_url"` // 配列ではなく、サムネイル1枚の文字列
	Status   string `json:"status"`
}

// IsValid バリデーション
func (req *ItemCreateRequest) IsValid() bool {
	return req.Name != "" && req.Price >= 0 && 10 > len(req.ImageURLs) && len(req.ImageURLs) > 0
}

// IsValid バリデーション
func (req *ItemUpdateRequest) IsValid() bool {
	if req.ItemID == "" {
		return false
	}
	if req.UserID == "" {
		return false
	}
	if req.Name == "" {
		return false
	}
	if req.Price <= 0 {
		return false
	}
	return true
}
