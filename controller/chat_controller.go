package controller

import (
	"db/middleware"
	"db/model"
	"db/usecase"
	"encoding/json"
	"net/http"
)

type ChatController struct {
	chatUsecase usecase.ChatUsecase
}

func NewChatController(u usecase.ChatUsecase) *ChatController {
	return &ChatController{chatUsecase: u}
}

// HandleGetOrCreateRoom : チャットルームを開始・取得 (POST /items/{item_id}/chat)
func (c *ChatController) HandleGetOrCreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	buyerID, err := middleware.GetUserIDFromContext(ctx) // ログインユーザー(購入者)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	itemID := r.PathValue("item_id")

	// リクエストボディから出品者IDを取得
	var req struct {
		SellerID string `json:"seller_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	room, err := c.chatUsecase.GetOrCreateChatRoom(ctx, itemID, buyerID, req.SellerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get or create room", err)
		return
	}

	respondJSON(w, http.StatusOK, room)
}

// HandleGetChatRoomList : チャットルーム一覧を取得 (GET /items/{item_id}/chat_rooms)
func (c *ChatController) HandleGetChatRoomList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemID := r.PathValue("item_id")
	rooms, err := c.chatUsecase.GetChatRoomList(ctx, itemID)
	if err != nil { respondError(w, http.StatusInternalServerError, "Failed to get chat list", err); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"rooms": rooms})
}

// HandleGetMessages : メッセージ一覧を取得 (GET /chats/{room_id}/messages)
func (c *ChatController) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	roomID := r.PathValue("room_id")

	messages, err := c.chatUsecase.GetMessages(ctx, roomID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get messages", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"messages": messages})
}

// HandleSendMessage : メッセージを送信 (POST /chats/{room_id}/messages)
func (c *ChatController) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	senderID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	roomID := r.PathValue("room_id")

	var req model.MessageSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := c.chatUsecase.SendMessage(ctx, roomID, senderID, req.Content); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send message", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "sent"})
}
