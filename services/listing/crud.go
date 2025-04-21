package listing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"carsawa/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateDealerListing creates a draft dealer listing and notifies the dealer.
func (s *listingService) CreateDealerListing(
	ctx context.Context,
	dealerHex string,
	listing models.Listing,
	price float64,
) (*models.Listing, error) {
	dealerID, err := s.helper.convertAndValidateID(dealerHex)
	if err != nil {
		return nil, err
	}
	if price <= 0 {
		return nil, errors.New("dealer listings require positive price")
	}

	toCreate := &models.Listing{
		ID:        primitive.NewObjectID(),
		Type:      models.ListingTypeDealer,
		Status:    models.ListingStatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DealerListing: models.DealerListing{
			DealerID: dealerID,
		},
		CarDetails: listing.CarDetails,
	}

	if err := s.validateCarDetails(ctx, *toCreate); err != nil {
		return nil, err
	}

	// persist draft
	newID, err := s.repo.CreateListing(ctx, toCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to create listing: %w", err)
	}
	lst, err := s.repo.GetListingByID(ctx, newID)
	if err != nil {
		return nil, err
	}

	// notify dealer
	s.notifyDealer(
		ctx,
		dealerID,
		models.NotificationTypeListingCreated,
		"Draft Listing Created",
		fmt.Sprintf("Your draft for %s %s has been saved.", lst.CarDetails.Make, lst.CarDetails.Model),
		map[string]interface{}{"listingID": lst.ID.Hex()},
	)

	return lst, nil
}

// GetListing fetches a listing and increments its view count asynchronously.
func (s *listingService) GetListing(
	ctx context.Context,
	listingID string,
) (*models.Listing, error) {
	if _, err := primitive.ObjectIDFromHex(listingID); err != nil {
		return nil, errors.New("invalid listing ID format")
	}
	lst, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	go s.repo.IncrementViews(context.Background(), listingID)
	return lst, nil
}

// UpdateListing applies arbitrary updates, then notifies any bidders if price changed.
func (s *listingService) UpdateListing(
	ctx context.Context,
	listingID string,
	updates map[string]interface{},
) (*models.Listing, error) {
	// validate ID
	if _, err := s.helper.convertAndValidateID(listingID); err != nil {
		return nil, err
	}

	// load existing for validation & diffing
	existing, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, err
	}

	// if carDetails provided, validate them
	if cdRaw, ok := updates["carDetails"]; ok {
		if err := s.validateAndUpdateCarDetails(existing, cdRaw); err != nil {
			return nil, err
		}
	}

	// no type changes allowed
	if _, ok := updates["type"]; ok {
		return nil, errors.New("cannot modify listing type")
	}

	// timestamp and persist
	updates["updatedAt"] = time.Now()
	if err := s.repo.UpdateListing(ctx, listingID, updates); err != nil {
		return nil, fmt.Errorf("failed to update listing: %w", err)
	}

	// reload post‑update
	updated, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, err
	}

	// if price changed, notify all dealers who have bids
	if cdRaw, ok := updates["carDetails"].(map[string]interface{}); ok {
		if newPrice, ok2 := cdRaw["price"].(float64); ok2 {
			for _, bid := range updated.UserListing.Bids {
				s.notifyDealer(
					ctx,
					bid.DealerID,
					models.NotificationTypeListingUpdated,
					"Listing Updated",
					fmt.Sprintf("Price for %s %s updated to %.2f.",
						updated.CarDetails.Make, updated.CarDetails.Model, newPrice,
					),
					map[string]interface{}{"listingID": listingID, "newPrice": newPrice},
				)
			}
		}
	}

	return updated, nil
}

// DeleteListing removes a listing (unless already accepted).
func (s *listingService) DeleteListing(
	ctx context.Context,
	listingID string,
) error {
	if _, err := primitive.ObjectIDFromHex(listingID); err != nil {
		return errors.New("invalid listing ID format")
	}
	lst, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return err
	}
	if lst.Status == models.ListingStatusAccepted {
		return errors.New("cannot delete accepted listings")
	}
	return s.repo.DeleteListing(ctx, listingID)
}

// PublishListing flips a draft to active, then notifies the dealer.
func (s *listingService) PublishListing(
	ctx context.Context,
	listingID, dealerHex string,
) (*models.Listing, error) {
	dealerID, err := primitive.ObjectIDFromHex(dealerHex)
	if err != nil {
		return nil, errors.New("invalid dealer ID format")
	}

	// update status via existing method
	published, err := s.UpdateListing(ctx, listingID, map[string]interface{}{
		"status": models.ListingStatusActive,
	})
	if err != nil {
		return nil, err
	}

	s.notifyDealer(
		ctx,
		dealerID,
		models.NotificationTypeListingPublished,
		"Listing Published",
		fmt.Sprintf("%s %s is now live.",
			published.CarDetails.Make, published.CarDetails.Model,
		),
		map[string]interface{}{"listingID": listingID},
	)

	return published, nil
}

// CloseListing marks a listing closed and notifies the closer.
func (s *listingService) CloseListing(
	ctx context.Context,
	listingID, ownerHex string,
	isDealer bool,
) error {
	ownerID, err := primitive.ObjectIDFromHex(ownerHex)
	if err != nil {
		return errors.New("invalid owner ID format")
	}

	lst, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return err
	}

	// auth checks
	if isDealer {
		if lst.DealerListing.DealerID != ownerID {
			return errors.New("unauthorized dealer operation")
		}
	} else {
		if lst.UserListing.UserID != ownerID {
			return errors.New("unauthorized user operation")
		}
	}

	// close status
	_, err = s.UpdateListing(ctx, listingID, map[string]interface{}{
		"status": models.ListingStatusClosed,
	})
	if err != nil {
		return err
	}

	// fire notification
	notifType := models.NotificationTypeListingClosed
	title := "Listing Closed"
	body := fmt.Sprintf("You closed %s %s.",
		lst.CarDetails.Make, lst.CarDetails.Model,
	)
	if isDealer {
		s.notifyDealer(ctx, ownerID, notifType, title, body, map[string]interface{}{"listingID": listingID})
	} else {
		s.notifyUser(ctx, ownerID, notifType, title, body, map[string]interface{}{"listingID": listingID})
	}

	return nil
}

func (s *listingService) SearchListings(
	ctx context.Context,
	filter models.ListingFilter,
	pagination models.Pagination,
) ([]models.Listing, error) {
	return s.repo.SearchListings(ctx, filter, pagination)
}
