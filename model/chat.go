package model

import "time"

// ChatRoom : チャットルーム構造体
type ChatRoom struct {
	Id        string    `json:"id"`
	ItemId    string    `json:"item_id"`
	BuyerId   string    `json:"buyer_id"`
	SellerId  string    `json:"seller_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Message : メッセージ構造体
type Message struct {
	Id         string    `json:"id"`
	ChatRoomId string    `json:"chat_room_id"`
	SenderId   string    `json:"sender_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

// リクエスト用
type MessageSendRequest struct {
	Content string `json:"content"`
}
