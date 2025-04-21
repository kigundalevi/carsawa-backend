package listing

import (
	listingRepo "carsawa/database/repository/listing"
	"carsawa/models"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type listingHelper struct {
	repo listingRepo.ListingRepository
}

func newListingHelper(repo listingRepo.ListingRepository) *listingHelper {
	return &listingHelper{repo: repo}
}

func (h *listingHelper) convertAndValidateID(id string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid ID format")
	}
	return objID, nil
}

func (h *listingHelper) isValidStatusTransition(oldStatus, newStatus models.ListingStatus) bool {
	transitions := map[models.ListingStatus][]models.ListingStatus{
		models.ListingStatusDraft:    {models.ListingStatusActive, models.ListingStatusClosed},
		models.ListingStatusActive:   {models.ListingStatusSold, models.ListingStatusClosed},
		models.ListingStatusOpen:     {models.ListingStatusAccepted, models.ListingStatusClosed},
		models.ListingStatusAccepted: {models.ListingStatusClosed},
	}

	allowed, exists := transitions[oldStatus]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}
	return false
}
