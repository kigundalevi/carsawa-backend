// listing_crud.go

package listingRepo

import (
	"carsawa/models"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateListing inserts a new listing into the database.
func (r *MongoListingsRepository) CreateListing(ctx context.Context, listing *models.Listing) (string, error) {
	if listing.ID.IsZero() {
		listing.ID = primitive.NewObjectID()
	}

	listing.CreatedAt = time.Now()
	listing.UpdatedAt = time.Now()

	switch listing.Type {
	case models.ListingTypeDealer:
		if listing.CarDetails.Price <= 0 {
			return "", errors.New("dealer listings require positive price")
		}
		listing.Status = models.ListingStatusDraft
		listing.UserListing.Bids = nil
	case models.ListingTypeUserBid:
		if listing.CarDetails.Price != 0 {
			return "", errors.New("user bid listings must not have price")
		}
		listing.Status = models.ListingStatusOpen
		listing.UserListing.Bids = make([]models.Bid, 0)
	default:
		return "", ErrInvalidType
	}

	_, err := r.listings.InsertOne(ctx, listing)
	if err != nil {
		return "", fmt.Errorf("failed to create listing: %w", err)
	}
	return listing.ID.Hex(), nil
}

// GetListingByID retrieves a listing by its ID.
func (r *MongoListingsRepository) GetListingByID(ctx context.Context, id string) (*models.Listing, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	var listing models.Listing
	err = r.listings.FindOne(ctx, bson.M{"_id": objID}).Decode(&listing)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	return &listing, nil
}

// UpdateListing updates specific fields of a listing.
func (r *MongoListingsRepository) UpdateListing(ctx context.Context, id string, updates map[string]interface{}) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	if _, ok := updates["type"]; ok {
		return errors.New("cannot modify listing type")
	}

	updates["updatedAt"] = time.Now()

	res, err := r.listings.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return fmt.Errorf("failed to update listing: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteListing removes a listing by its ID.
func (r *MongoListingsRepository) DeleteListing(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	res, err := r.listings.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete listing: %w", err)
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MongoListingsRepository) DeleteListingsByDealerID(ctx context.Context, dealerID string) error {
	filter := bson.M{"dealerId": dealerID}
	res, err := r.listings.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
	}
	return nil
}
