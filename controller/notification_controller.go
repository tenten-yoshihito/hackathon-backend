package controller

import (
	"db/dao"
	"db/middleware"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type NotificationController struct {
	notificationDAO dao.NotificationDAO
}

func NewNotificationController(notificationDAO dao.NotificationDAO) *NotificationController {
	return &NotificationController{
		notificationDAO: notificationDAO,
	}
}

// HandleGetNotifications : 通知一覧を取得
func (c *NotificationController) HandleGetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR NOTIFICATION] Failed to get user ID: %v\n", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// デフォルトは50件
	limit := 50

	notifications, err := c.notificationDAO.GetUserNotifications(r.Context(), userID, limit)
	if err != nil {
		log.Printf("Failed to get notifications: %v\n", err)
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	json.NewEncoder(w).Encode(notifications)
}

// HandleGetUnreadCount : 未読通知数を取得
func (c *NotificationController) HandleGetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	count, err := c.notificationDAO.GetUnreadCount(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get unread count: %v\n", err)
		http.Error(w, "Failed to get unread count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	json.NewEncoder(w).Encode(map[string]int{"unread_count": count})
}

// HandleMarkAsRead : 通知を既読にする
func (c *NotificationController) HandleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Failed to get user ID: %v\n", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// URLから通知IDを取得（/notifications/xxx/read）
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}
	notificationId := pathParts[2]

	err = c.notificationDAO.MarkAsRead(r.Context(), notificationId, userID)
	if err != nil {
		log.Printf("Failed to mark notification as read: %v\n", err)
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleMarkAllAsRead : すべての通知を既読にする
func (c *NotificationController) HandleMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Failed to get user ID: %v\n", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.notificationDAO.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to mark all notifications as read: %v\n", err)
		http.Error(w, "Failed to mark all notifications as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
