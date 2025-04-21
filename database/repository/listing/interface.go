package listingRepo

import (
	"carsawa/models"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// ListingType distinguishes between dealer and user listings
type ListingType string

const (
	DealerListingType  ListingType = "dealer"
	UserBidListingType ListingType = "user_bid"
)

var (
	ErrNotFound           = errors.New("listing not found")
	ErrInvalidID          = errors.New("invalid listing ID")
	ErrInvalidType        = errors.New("invalid listing type")
	ErrBidConflict        = errors.New("bid conflict occurred")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrUnauthorizedAction = errors.New("unauthorized listing action")
)

type ListingRepository interface {
	CreateListing(ctx context.Context, listing *models.Listing) (string, error)
	GetListingByID(ctx context.Context, id string) (*models.Listing, error)
	UpdateListing(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteListing(ctx context.Context, id string) error
	AddBid(ctx context.Context, listingID string, bid models.Bid) error
	AcceptBid(ctx context.Context, listingID, bidID string) error
	SearchListings(ctx context.Context, filter models.ListingFilter, pagination models.Pagination) ([]models.Listing, error)
	IncrementViews(ctx context.Context, listingID string) error
	DeleteListingsByDealerID(ctx context.Context, dealerID string) error

	// Feed operations
	GetActiveListings(ctx context.Context, filter models.ListingFilter, pagination models.Pagination) ([]models.Listing, error)
	GetFeaturedListings(ctx context.Context, limit int) ([]models.Listing, error)
	GetPromotions(ctx context.Context) ([]models.Promotion, error)
	GetBanners(ctx context.Context) ([]models.Banner, error)
	TextSearch(ctx context.Context, query string, pagination models.Pagination) ([]models.Listing, error)
	GetSearchSuggestions(ctx context.Context, query string) ([]models.SearchSuggestion, error)
	RecordListingView(ctx context.Context, listingID string) error
	RecordSearchQuery(ctx context.Context, query string, filters models.ListingFilter) error
}

type MongoListingsRepository struct {
	listings *mongo.Collection
}

func NewMongoListingsRepository(db *mongo.Database) *MongoListingsRepository {
	return &MongoListingsRepository{
		listings: db.Collection("listings"),
	}
}
