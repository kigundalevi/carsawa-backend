package dealer

import (
	"carsawa/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (s *dealerService) UpdateDealer(ctx context.Context, id string, updates bson.M) (*models.Dealer, error) {
	// Protect immutable fields
	delete(updates, "_id")
	delete(updates, "email")
	delete(updates, "createdAt")
	delete(updates, "slug") // Slug should be updated through separate endpoint if needed

	// Set updatedAt
	updates["updatedAt"] = time.Now()

	if err := s.repo.UpdateDealer(id, bson.M{"$set": updates}); err != nil {
		return nil, fmt.Errorf("failed to update dealer: %w", err)
	}

	return s.GetDealer(ctx, id)
}

// Update the dealer service methods to handle projections
func (s *dealerService) GetDealer(ctx context.Context, id string) (*models.Dealer, error) {
	isFullAccess := getAccessLevelFromContext(ctx)
	projection := buildDealerProjection(isFullAccess)

	dealer, err := s.repo.GetDealerByIDWithProjection(id, projection)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDealerNotFound, err)
	}
	return dealer, nil
}

func (s *dealerService) GetDealerByEmail(ctx context.Context, email string) (*models.Dealer, error) {
	isFullAccess := getAccessLevelFromContext(ctx)
	projection := buildDealerProjection(isFullAccess)

	dealer, err := s.repo.GetDealerByEmailWithProjection(email, projection)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDealerNotFound, err)
	}
	return dealer, nil
}

func (s *dealerService) ListDealers(ctx context.Context) ([]models.Dealer, error) {
	isFullAccess := getAccessLevelFromContext(ctx)
	projection := buildDealerProjection(isFullAccess)

	dealers, err := s.repo.GetAllDealersWithProjection(projection)
	if err != nil {
		return nil, fmt.Errorf("failed to list dealers: %w", err)
	}
	return dealers, nil
}

// Helper functions
func getAccessLevelFromContext(ctx context.Context) bool {
	if val := ctx.Value("isDealerFullAccess"); val != nil {
		return val.(bool)
	}
	return false // Default to partial access
}

func buildDealerProjection(fullAccess bool) bson.M {
	if fullAccess {
		return bson.M{
			"security": 0, // Exclude security field
		}
	}
	// Partial access projection
	return bson.M{
		"_id":     1,
		"profile": 1,
		"store":   1,
	}
}

func (s *dealerService) DeleteDealer(ctx context.Context, id string) error {
	// Delete related listings first
	if err := s.listingsRepo.DeleteListingsByDealerID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete listings for dealer: %w", err)
	}

	if err := s.repo.DeleteDealer(id); err != nil {
		return fmt.Errorf("failed to delete dealer: %w", err)
	}

	return nil
}
