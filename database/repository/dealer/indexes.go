package dealerRepo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ensureIndexes creates indexes for frequently used fields in queries.
func (r *mongoDealerRepo) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		// Unique index on dealer's "id".
		{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
		// Unique index on the provider's email stored in "profile.dealerName".
		{Keys: bson.D{{Key: "profile.slug", Value: 1}}, Options: options.Index().SetUnique(true)},
		// Unique index on the provider's email stored in "profile.dealerName".
		{Keys: bson.D{{Key: "profile.dealerName", Value: 1}}, Options: options.Index().SetUnique(true)},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	return nil
}
