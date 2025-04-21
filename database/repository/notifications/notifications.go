package notificationsRepo

import (
	"carsawa/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *MongoNotificationRepository) CreateNotification(ctx context.Context, n *models.Notification) error {
	n.ID = primitive.NewObjectID()
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	n.Read = false

	_, err := r.collection.InsertOne(ctx, n)
	return err
}

func (r *MongoNotificationRepository) FindByRecipient(ctx context.Context, recipientID primitive.ObjectID, target models.NotificationTarget, page, limit int) ([]models.Notification, error) {
	filter := bson.M{"recipient": recipientID, "target": target}
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []models.Notification
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *MongoNotificationRepository) MarkOneRead(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"read": true, "updatedAt": time.Now()}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoNotificationRepository) MarkAllRead(ctx context.Context, recipientID primitive.ObjectID, target models.NotificationTarget) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"recipient": recipientID, "target": target, "read": false},
		bson.M{"$set": bson.M{"read": true, "updatedAt": time.Now()}},
	)
	return err
}
