package dao

import (
	"context"
	"crypto/rand"
	"database/sql"
	"db/model"
	"fmt"
	"time"

	"github.com/oklog/ulid"
)

type NotificationDAO interface {
	CreateNotification(ctx context.Context, notification *model.Notification) error
	GetUserNotifications(ctx context.Context, userId string, limit int) ([]model.Notification, error)
	GetUnreadCount(ctx context.Context, userId string) (int, error)
	MarkAsRead(ctx context.Context, notificationId string, userId string) error
	MarkAllAsRead(ctx context.Context, userId string) error
}

type notificationDao struct {
	DB *sql.DB
}

func NewNotificationDAO(db *sql.DB) NotificationDAO {
	return &notificationDao{DB: db}
}

// CreateNotification : 通知を作成
func (dao *notificationDao) CreateNotification(ctx context.Context, notification *model.Notification) error {
	// IDとCreatedAtが設定されていない場合は自動生成
	if notification.Id == "" {
		t := time.Now()
		entropy := ulid.Monotonic(rand.Reader, 0)
		notification.Id = ulid.MustNew(ulid.Timestamp(t), entropy).String()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	query := `INSERT INTO notifications (id, user_id, type, item_id, item_name, message, is_read, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := dao.DB.ExecContext(ctx, query,
		notification.Id,
		notification.UserId,
		notification.Type,
		notification.ItemId,
		notification.ItemName,
		notification.Message,
		notification.IsRead,
		notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetUserNotifications : ユーザーの通知一覧を取得
func (dao *notificationDao) GetUserNotifications(ctx context.Context, userId string, limit int) ([]model.Notification, error) {
	query := `SELECT id, user_id, type, item_id, item_name, message, is_read, created_at
	          FROM notifications
	          WHERE user_id = ?
	          ORDER BY created_at DESC
	          LIMIT ?`

	rows, err := dao.DB.QueryContext(ctx, query, userId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]model.Notification, 0)
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.Id, &n.UserId, &n.Type, &n.ItemId, &n.ItemName, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return notifications, nil
}

// GetUnreadCount : 未読通知数を取得
func (dao *notificationDao) GetUnreadCount(ctx context.Context, userId string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`

	var count int
	err := dao.DB.QueryRowContext(ctx, query, userId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// MarkAsRead : 通知を既読にする
func (dao *notificationDao) MarkAsRead(ctx context.Context, notificationId string, userId string) error {
	query := `UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?`

	result, err := dao.DB.ExecContext(ctx, query, notificationId, userId)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or unauthorized")
	}

	return nil
}

// MarkAllAsRead : すべての通知を既読にする
func (dao *notificationDao) MarkAllAsRead(ctx context.Context, userId string) error {
	query := `UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0`

	_, err := dao.DB.ExecContext(ctx, query, userId)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}
