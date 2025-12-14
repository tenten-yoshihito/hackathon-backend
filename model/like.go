package model

import "time"

// Like : いいねの構造体
type Like struct {
	UserID    string    `json:"user_id"`
	ItemID    string    `json:"item_id"`
	CreatedAt time.Time `json:"created_at"`
}
