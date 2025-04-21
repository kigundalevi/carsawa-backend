package listing

import (
	"carsawa/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *listingService) CreateUserBidListing(
	ctx context.Context,
	userHex string,
	listing models.Listing,
) (*models.Listing, error) {
	// 1) Validate & build
	userID, err := s.helper.convertAndValidateID(userHex)
	if err != nil {
		return nil, err
	}

	toCreate := &models.Listing{
		ID:        primitive.NewObjectID(),
		Type:      models.ListingTypeUserBid,
		Status:    models.ListingStatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserListing: models.UserListing{
			UserID:      userID,
			Bids:        []models.Bid{},
			AcceptedBid: nil,
		},
		CarDetails: listing.CarDetails,
	}

	if err := s.validateCarDetails(ctx, *toCreate); err != nil {
		return nil, err
	}

	// 2) Persist & reload
	newID, err := s.repo.CreateListing(ctx, toCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to create user‑bid listing: %w", err)
	}
	lst, err := s.repo.GetListingByID(ctx, newID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch new listing: %w", err)
	}

	// 3) Notify the user that their listing is live
	s.notifyUser(
		ctx,
		userID,
		models.NotificationTypeListingCreated,
		"Your Bid Listing Is Live",
		fmt.Sprintf("Your listing for %s %s is now open for dealer bids.",
			lst.CarDetails.Make, lst.CarDetails.Model),
		map[string]interface{}{"listingID": lst.ID.Hex()},
	)

	return lst, nil
}

func (s *listingService) AddBid(
	ctx context.Context,
	listingID string,
	bid models.Bid,
) (*models.Listing, error) {
	// 1) Basic validation & repo call
	if _, err := primitive.ObjectIDFromHex(listingID); err != nil {
		return nil, fmt.Errorf("invalid listing ID: %w", err)
	}
	if err := s.repo.AddBid(ctx, listingID, bid); err != nil {
		return nil, fmt.Errorf("failed to add bid: %w", err)
	}

	// 2) Reload the updated listing
	lst, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("reload listing: %w", err)
	}

	// 3) Fetch dealer info (for friendly message)
	dealer, err := s.dealer.GetDealer(ctx, bid.DealerID.Hex())
	dealerName := "A dealer"
	if err == nil {
		dealerName = dealer.Profile.DealerName
	}

	// 4) Build notification payload
	title := "New Bid Placed"
	body := fmt.Sprintf("%s placed a bid of %.2f on your %s %s",
		dealerName, bid.Offer, lst.CarDetails.Make, lst.CarDetails.Model,
	)
	data := map[string]interface{}{
		"listingID": listingID,
		"offer":     bid.Offer,
	}

	// 5) Fire‑and‑forget notify
	s.notifyUser(ctx, lst.UserListing.UserID, models.NotificationTypeBidPlaced, title, body, data)

	return lst, nil
}

func (s *listingService) AcceptBid(
	ctx context.Context,
	listingID, bidID, userHex string,
) (*models.Listing, error) {
	// 1) Validate & accept bid in repo
	if _, err := primitive.ObjectIDFromHex(listingID); err != nil {
		return nil, fmt.Errorf("invalid listing ID: %w", err)
	}
	if _, err := primitive.ObjectIDFromHex(bidID); err != nil {
		return nil, fmt.Errorf("invalid bid ID: %w", err)
	}
	if err := s.repo.AcceptBid(ctx, listingID, bidID); err != nil {
		return nil, fmt.Errorf("failed to accept bid: %w", err)
	}

	// 2) Reload the listing with updated accepted bid
	lst, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("reload listing: %w", err)
	}

	// 3) Find the accepted bid object
	var accepted models.Bid
	bidOID, _ := primitive.ObjectIDFromHex(bidID)
	for _, b := range lst.UserListing.Bids {
		if b.ID == bidOID {
			accepted = b
			break
		}
	}

	// 4) Fetch user info for friendly message
	user, err := s.user.GetUserByID(ctx, userHex)
	username := "The listing owner"
	if err == nil {
		username = user.Username
	}

	// 5) Build and send notification to dealer
	title := "Congratulations! Your Bid Was Accepted"
	body := fmt.Sprintf("%s accepted your bid of %.2f on their %s %s",
		username, accepted.Offer, lst.CarDetails.Make, lst.CarDetails.Model,
	)
	data := map[string]interface{}{
		"listingID": listingID,
		"offer":     accepted.Offer,
	}

	s.notifyDealer(ctx, accepted.DealerID, models.NotificationTypeBidAccepted, title, body, data)

	return lst, nil
}
