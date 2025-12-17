package usecase

import (
	"context"
	"crypto/rand"
	"db/dao"
	"db/model"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid"
)

type ChatUsecase interface {
	// ルーム関連
	GetOrCreateChatRoom(ctx context.Context, itemID, buyerID, sellerID string) (*model.ChatRoom, error)
	GetChatRoomList(ctx context.Context, itemID string) ([]model.ChatRoomInfo, error)

	// メッセージ関連
	SendMessage(ctx context.Context, roomID, senderID, content string) error
	GetMessages(ctx context.Context, roomID string) ([]model.Message, error)
}

type chatUsecase struct {
	chatDAO         dao.ChatDAO
	itemDAO         dao.ItemDAO
	notificationDAO dao.NotificationDAO
}

func NewChatUsecase(chatDAO dao.ChatDAO, itemDAO dao.ItemDAO, notificationDAO dao.NotificationDAO) ChatUsecase {
	return &chatUsecase{
		chatDAO:         chatDAO,
		itemDAO:         itemDAO,
		notificationDAO: notificationDAO,
	}
}

// GetOrCreateChatRoom :チャットルームがあれば取得、なければ作成して返す
func (u *chatUsecase) GetOrCreateChatRoom(ctx context.Context, itemID, buyerID, sellerID string) (*model.ChatRoom, error) {
	// 1. 既存のルームがあるか探す
	existingRoom, err := u.chatDAO.GetChatRoom(ctx, itemID, buyerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing room: %w", err)
	}
	if existingRoom != nil {
		return existingRoom, nil // あればそれを返す
	}

	// 2. なければ新規作成
	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	newID := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	newRoom := &model.ChatRoom{
		Id:        newID,
		ItemId:    itemID,
		BuyerId:   buyerID,
		SellerId:  sellerID,
		CreatedAt: t,
	}

	if err := u.chatDAO.CreateChatRoom(ctx, newRoom); err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	return newRoom, nil
}

// GetChatRoomList :商品IDからチャットルーム一覧を取得
func (u *chatUsecase) GetChatRoomList(ctx context.Context, itemID string) ([]model.ChatRoomInfo, error) {
	return u.chatDAO.GetChatRoomsByItemID(ctx, itemID)
}

// SendMessage :メッセージを送信
func (u *chatUsecase) SendMessage(ctx context.Context, roomID, senderID, content string) error {
	if content == "" {
		return fmt.Errorf("message content is empty")
	}

	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	newID := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	msg := &model.Message{
		Id:         newID,
		ChatRoomId: roomID,
		SenderId:   senderID,
		Content:    content,
		CreatedAt:  t,
	}

	if err := u.chatDAO.SaveMessage(ctx, msg); err != nil {
		return err
	}

	// チャットルーム情報を取得
	room, err := u.chatDAO.GetChatRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Warning: failed to get chat room: %v\n", err)
		return nil
	}

	// 商品情報を取得
	item, err := u.itemDAO.GetItem(ctx, room.ItemId)
	if err != nil {
		log.Printf("Warning: failed to get item: %v\n", err)
		return nil
	}

	// コメント通知を作成（相手に通知）
	var recipientID string
	switch senderID {
	case room.BuyerId:
		recipientID = room.SellerId
	case room.SellerId:
		recipientID = room.BuyerId
	default:
		return nil
	}

	notificationID := ulid.MustNew(ulid.Timestamp(t), entropy).String()
	notification := &model.Notification{
		Id:        notificationID,
		UserId:    recipientID,
		Type:      "comment",
		ItemId:    room.ItemId,
		ItemName:  item.Name,
		Message:   fmt.Sprintf("%sにコメントがつきました", item.Name),
		IsRead:    false,
		CreatedAt: t,
	}

	if err := u.notificationDAO.CreateNotification(ctx, notification); err != nil {
		log.Printf("Warning: failed to create notification: %v\n", err)
	}

	return nil
}

// GetMessages :メッセージ履歴を取得
func (u *chatUsecase) GetMessages(ctx context.Context, roomID string) ([]model.Message, error) {
	return u.chatDAO.GetMessages(ctx, roomID)
}
