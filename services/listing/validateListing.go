package listing

import (
	"carsawa/models"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type listingValidator struct {
	verifier *NHTSAVerifier
}

func newListingValidator(verifier *NHTSAVerifier) *listingValidator {
	return &listingValidator{verifier: verifier}
}

func (s *listingService) validateCarDetails(
	ctx context.Context,
	listing models.Listing,
) error {
	cd := listing.CarDetails

	if cd.VIN == "" {
		return errors.New("vIN is required")
	}
	if cd.Make == "" {
		return errors.New("make is required")
	}
	if cd.Model == "" {
		return errors.New("model is required")
	}

	currentYear := time.Now().Year()
	if cd.Year < 1886 || cd.Year > currentYear+1 {
		return errors.New("invalid manufacturing year")
	}

	switch listing.Type {
	case models.ListingTypeDealer:
		if cd.Price <= 0 {
			return errors.New("price must be positive for dealer listings")
		}
	case models.ListingTypeUserBid:
		if cd.Price != 0 {
			return errors.New("user bid listings cannot have fixed price")
		}
	default:
		return errors.New("invalid listing type")
	}

	// verify VIN and decode official details
	res, valid, err := s.verifier.VerifyVIN(ctx, cd.VIN)
	if err != nil {
		return fmt.Errorf("VIN verification failed: %w", err)
	}
	if !valid {
		return errors.New("invalid VIN")
	}

	// compare Make
	if !strings.EqualFold(res.Make, cd.Make) {
		return fmt.Errorf("provided make %q does not match registry %q", cd.Make, res.Make)
	}
	// compare Model
	if !strings.EqualFold(res.Model, cd.Model) {
		return fmt.Errorf("provided model %q does not match registry %q", cd.Model, res.Model)
	}
	// compare Year
	regYear, err := strconv.Atoi(res.ModelYear)
	if err != nil {
		return fmt.Errorf("invalid year from registry: %w", err)
	}
	if regYear != cd.Year {
		return fmt.Errorf("provided year %d does not match registry %d", cd.Year, regYear)
	}

	return nil
}

func (s *listingService) validateAndUpdateCarDetails(existing *models.Listing, carDetailsRaw interface{}) error {
	updatedCarDetails := existing.CarDetails
	if carDetailsMap, ok := carDetailsRaw.(map[string]interface{}); ok {
		if makeVal, ok := carDetailsMap["make"].(string); ok {
			updatedCarDetails.Make = makeVal
		}
		if modelVal, ok := carDetailsMap["model"].(string); ok {
			updatedCarDetails.Model = modelVal
		}
		if yearVal, ok := carDetailsMap["year"].(int); ok {
			updatedCarDetails.Year = yearVal
		}
		if vinVal, ok := carDetailsMap["vin"].(string); ok {
			updatedCarDetails.VIN = vinVal
		}
		if priceVal, ok := carDetailsMap["price"].(float64); ok {
			updatedCarDetails.Price = priceVal
		}
	}

	tempListing := *existing
	tempListing.CarDetails = updatedCarDetails
	return s.validateCarDetails(context.Background(), tempListing)
}
