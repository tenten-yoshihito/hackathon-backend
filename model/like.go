package model

import "time"

// Like represents a user's like on an item
type Like struct {
	UserID    string    `json:"user_id"`
	ItemID    string    `json:"item_id"`
	CreatedAt time.Time `json:"created_at"`
}
