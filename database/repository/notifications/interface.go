package notificationsRepo

import (
	"carsawa/models"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoNotificationRepository struct {
	collection *mongo.Collection
}

type NotificationRepository interface {
	CreateNotification(ctx context.Context, n *models.Notification) error
	FindByRecipient(ctx context.Context, recipientID primitive.ObjectID, target models.NotificationTarget, page, limit int) ([]models.Notification, error)
	MarkOneRead(ctx context.Context, notificationID primitive.ObjectID) error
	MarkAllRead(ctx context.Context, recipientID primitive.ObjectID, target models.NotificationTarget) error
}

func NewMongoNotificationRepository(db *mongo.Database) *MongoNotificationRepository {
	return &MongoNotificationRepository{
		collection: db.Collection("notifications"),
	}
}
