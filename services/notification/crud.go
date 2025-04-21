package notification

import (
	"carsawa/models"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *notificationService) CreateUserNotification(
	ctx context.Context,
	userID string,
	nt models.NotificationType,
	title, body string,
	data map[string]interface{},
) error {
	return s.sendNotification(ctx, models.NotificationTargetUser, userID, nt, title, body, data)
}

func (s *notificationService) CreateDealerNotification(
	ctx context.Context,
	dealerID string,
	nt models.NotificationType,
	title, body string,
	data map[string]interface{},
) error {
	return s.sendNotification(ctx, models.NotificationTargetDealer, dealerID, nt, title, body, data)
}

func (s *notificationService) sendNotification(
	ctx context.Context,
	target models.NotificationTarget,
	recipientHex string,
	nt models.NotificationType,
	title, body string,
	data map[string]interface{},
) error {
	if title == "" || body == "" {
		return ErrInvalidNotification
	}

	recipientID, err := primitive.ObjectIDFromHex(recipientHex)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidRecipientID, recipientHex)
	}

	notif := &models.Notification{
		Recipient: recipientID,
		Target:    target,
		Type:      nt,
		Title:     title,
		Body:      body,
		Data:      data,
		Read:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateNotification(ctx, notif); err != nil {
		return fmt.Errorf("sendNotification: failed to store: %w", err)
	}
	return nil
}

func (s *notificationService) GetUserNotifications(
	ctx context.Context, userID string, page, limit int,
) ([]models.Notification, error) {
	return s.fetchNotifications(ctx, models.NotificationTargetUser, userID, page, limit)
}

func (s *notificationService) GetDealerNotifications(
	ctx context.Context, dealerID string, page, limit int,
) ([]models.Notification, error) {
	return s.fetchNotifications(ctx, models.NotificationTargetDealer, dealerID, page, limit)
}

func (s *notificationService) fetchNotifications(
	ctx context.Context,
	target models.NotificationTarget,
	recipientHex string,
	page, limit int,
) ([]models.Notification, error) {
	recipientID, err := primitive.ObjectIDFromHex(recipientHex)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRecipientID, recipientHex)
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.FindByRecipient(ctx, recipientID, target, page, limit)
}

func (s *notificationService) MarkNotificationRead(
	ctx context.Context, notificationHex string,
) error {
	notifID, err := primitive.ObjectIDFromHex(notificationHex)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidRecipientID, notificationHex)
	}
	err = s.repo.MarkOneRead(ctx, notifID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotificationNotFound
	}
	return err
}

func (s *notificationService) MarkAllUserNotificationsRead(
	ctx context.Context, userID string,
) error {
	return s.markAllRead(ctx, models.NotificationTargetUser, userID)
}

func (s *notificationService) MarkAllDealerNotificationsRead(
	ctx context.Context, dealerID string,
) error {
	return s.markAllRead(ctx, models.NotificationTargetDealer, dealerID)
}

func (s *notificationService) markAllRead(
	ctx context.Context,
	target models.NotificationTarget,
	recipientHex string,
) error {
	recipientID, err := primitive.ObjectIDFromHex(recipientHex)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidRecipientID, recipientHex)
	}
	return s.repo.MarkAllRead(ctx, recipientID, target)
}
