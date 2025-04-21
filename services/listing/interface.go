package listing

import (
	listingRepo "carsawa/database/repository/listing"
	"carsawa/models"
	"carsawa/services/dealer"
	"carsawa/services/notification"
	"carsawa/services/user"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ListingService interface {
	CreateDealerListing(ctx context.Context, dealerID string, car models.Listing, price float64) (*models.Listing, error)
	CreateUserBidListing(ctx context.Context, userID string, car models.Listing) (*models.Listing, error)
	GetListing(ctx context.Context, id string) (*models.Listing, error)
	UpdateListing(ctx context.Context, id string, updates map[string]interface{}) (*models.Listing, error)
	DeleteListing(ctx context.Context, id string) error
	AddBid(ctx context.Context, listingID string, bid models.Bid) (*models.Listing, error)
	AcceptBid(ctx context.Context, listingID, bidID, userID string) (*models.Listing, error)
	PublishListing(ctx context.Context, listingID, dealerID string) (*models.Listing, error)
	CloseListing(ctx context.Context, listingID, ownerID string, isDealer bool) error
	SearchListings(ctx context.Context, filter models.ListingFilter, pagination models.Pagination) ([]models.Listing, error)
	GetFeed(ctx context.Context, filter models.ListingFilter, pagination models.Pagination) (*models.FeedResponse, error)
	Search(ctx context.Context, query string, filters models.ListingFilter, pagination models.Pagination) (*models.SearchResult, error)
}

type listingService struct {
	repo      listingRepo.ListingRepository
	user      user.UserService
	dealer    dealer.DealerService
	verifier  *NHTSAVerifier
	validator *listingValidator
	helper    *listingHelper
	notifier  notification.NotificationService
}

type FeedResponse struct {
	Listings   []models.Listing   `json:"listings"`
	Promotions []models.Promotion `json:"promotions"`
	Banners    []models.Banner    `json:"banners"`
}

type SearchResponse struct {
	Results     []models.Listing     `json:"results"`
	Suggestions []string             `json:"suggestions"`
	Filters     models.ListingFilter `json:"filters"`
}

func NewListingService(
	repo listingRepo.ListingRepository,
	notifSvc notification.NotificationService,
	user user.UserService,
	dealer dealer.DealerService,
) ListingService {
	verifier := NewNHTSAVerifier()
	return &listingService{
		repo:      repo,
		user:      user,
		verifier:  verifier,
		validator: newListingValidator(verifier),
		helper:    newListingHelper(repo),
		notifier:  notifSvc,
	}
}

type idHelper interface {
	convertAndValidateID(string) (primitive.ObjectID, error)
}
