package listingRepo

import (
	"carsawa/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddBid adds a new bid to a user bid listing.
func (r *MongoListingsRepository) AddBid(ctx context.Context, listingID string, bid models.Bid) error {
	objID, err := primitive.ObjectIDFromHex(listingID)
	if err != nil {
		return ErrInvalidID
	}

	bid.ID = primitive.NewObjectID()
	bid.CreatedAt = time.Now()
	bid.UpdatedAt = time.Now()

	update := bson.M{
		"$push": bson.M{"bids": bid},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	res, err := r.listings.UpdateOne(
		ctx,
		bson.M{
			"_id":    objID,
			"type":   models.ListingTypeUserBid,
			"status": models.ListingStatusOpen,
		},
		update,
	)

	if err != nil {
		return fmt.Errorf("failed to add bid: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// AcceptBid sets a specific bid as accepted and updates the listing status.
func (r *MongoListingsRepository) AcceptBid(ctx context.Context, listingID, bidID string) error {
	listingObjID, err := primitive.ObjectIDFromHex(listingID)
	if err != nil {
		return ErrInvalidID
	}

	bidObjID, err := primitive.ObjectIDFromHex(bidID)
	if err != nil {
		return ErrInvalidID
	}

	session, err := r.listings.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		var listing models.Listing
		err := r.listings.FindOne(sessCtx, bson.M{
			"_id":      listingObjID,
			"bids._id": bidObjID,
			"status":   models.ListingStatusOpen,
		}).Decode(&listing)

		if err != nil {
			return nil, ErrNotFound
		}

		update := bson.M{
			"$set": bson.M{
				"status":      models.ListingStatusAccepted,
				"acceptedBid": bidObjID,
				"updatedAt":   time.Now(),
			},
		}

		res, err := r.listings.UpdateByID(sessCtx, listingObjID, update)
		if err != nil {
			return nil, fmt.Errorf("failed to accept bid: %w", err)
		}
		if res.MatchedCount == 0 {
			return nil, ErrNotFound
		}

		return nil, nil
	})

	return err
}
