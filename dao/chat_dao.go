package dao

import (
	"context"
	"database/sql"
	"db/model"
	"errors"
	"fmt"
)

type ChatDAO interface {
	CreateChatRoom(ctx context.Context, room *model.ChatRoom) error
	GetChatRoom(ctx context.Context, itemID, buyerID string) (*model.ChatRoom, error)
	GetChatRoomByID(ctx context.Context, roomID string) (*model.ChatRoom, error)
	SaveMessage(ctx context.Context, msg *model.Message) error
	GetMessages(ctx context.Context, roomID string) ([]model.Message, error)
}

type chatDao struct {
	DB *sql.DB
}

func NewChatDao(db *sql.DB) ChatDAO {
	return &chatDao{DB: db}
}

// チャットルーム作成
func (dao *chatDao) CreateChatRoom(ctx context.Context, room *model.ChatRoom) error {
	query := `INSERT INTO chat_rooms (id, item_id, buyer_id, seller_id, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := dao.DB.ExecContext(ctx, query, room.Id, room.ItemId, room.BuyerId, room.SellerId, room.CreatedAt)
	if err != nil {
		return fmt.Errorf("create chat room failed: %w", err)
	}
	return nil
}

// 既存ルームの取得 (商品IDと購入者IDで検索)
func (dao *chatDao) GetChatRoom(ctx context.Context, itemID, buyerID string) (*model.ChatRoom, error) {
	query := `SELECT id, item_id, buyer_id, seller_id, created_at FROM chat_rooms WHERE item_id = ? AND buyer_id = ?`
	row := dao.DB.QueryRowContext(ctx, query, itemID, buyerID)

	var room model.ChatRoom
	if err := row.Scan(&room.Id, &room.ItemId, &room.BuyerId, &room.SellerId, &room.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // 存在しない場合はnil
		}
		return nil, fmt.Errorf("get chat room failed: %w", err)
	}
	return &room, nil
}

// ルームIDから取得
func (dao *chatDao) GetChatRoomByID(ctx context.Context, roomID string) (*model.ChatRoom, error) {
	query := `SELECT id, item_id, buyer_id, seller_id, created_at FROM chat_rooms WHERE id = ?`
	row := dao.DB.QueryRowContext(ctx, query, roomID)
	var room model.ChatRoom
	if err := row.Scan(&room.Id, &room.ItemId, &room.BuyerId, &room.SellerId, &room.CreatedAt); err != nil {
		return nil, fmt.Errorf("get chat room by id failed: %w", err)
	}
	return &room, nil
}

// メッセージ保存
func (dao *chatDao) SaveMessage(ctx context.Context, msg *model.Message) error {
	query := `INSERT INTO messages (id, chat_room_id, sender_id, content, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := dao.DB.ExecContext(ctx, query, msg.Id, msg.ChatRoomId, msg.SenderId, msg.Content, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("save message failed: %w", err)
	}
	return nil
}

// メッセージ一覧取得
func (dao *chatDao) GetMessages(ctx context.Context, roomID string) ([]model.Message, error) {
	query := `SELECT id, chat_room_id, sender_id, content, created_at FROM messages WHERE chat_room_id = ? ORDER BY created_at ASC`
	rows, err := dao.DB.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("get messages failed: %w", err)
	}
	defer rows.Close()

	var msgs []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.Id, &m.ChatRoomId, &m.SenderId, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message failed: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}