package listingRepo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *MongoListingsRepository) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "carDetails.make", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "carDetails.mileage", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "carDetails.model", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "carDetails.year", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("status_createdAt"),
		},
		{
			Keys: bson.D{
				{Key: "carDetails.make", Value: "text"},
				{Key: "carDetails.model", Value: "text"},
			},
			Options: options.Index().SetName("car_text_search"),
		},
	}

	_, err := r.listings.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	return nil
}
