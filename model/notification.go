package model

import "time"

// Notification : 通知
type Notification struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	Type      string    `json:"type"` // "purchase" or "comment"
	ItemId    string    `json:"item_id"`
	ItemName  string    `json:"item_name"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
