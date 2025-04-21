package notification

import (
	"context"
	"errors"

	notificationsRepo "carsawa/database/repository/notifications"
	"carsawa/models"
)

var (
	ErrInvalidRecipientID   = errors.New("invalid recipient ID format")
	ErrInvalidNotification  = errors.New("notification must include title and body")
	ErrNotificationNotFound = errors.New("notification not found")
)

type NotificationService interface {
	CreateUserNotification(ctx context.Context, userID string, nt models.NotificationType, title, body string, data map[string]interface{}) error
	CreateDealerNotification(ctx context.Context, dealerID string, nt models.NotificationType, title, body string, data map[string]interface{}) error

	GetUserNotifications(ctx context.Context, userID string, page, limit int) ([]models.Notification, error)
	GetDealerNotifications(ctx context.Context, dealerID string, page, limit int) ([]models.Notification, error)

	MarkNotificationRead(ctx context.Context, notificationID string) error
	MarkAllUserNotificationsRead(ctx context.Context, userID string) error
	MarkAllDealerNotificationsRead(ctx context.Context, dealerID string) error
}

type notificationService struct {
	repo notificationsRepo.NotificationRepository
}

func NewNotificationService(repo notificationsRepo.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}
