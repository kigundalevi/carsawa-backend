// File: interface.go
package dealerRepo

import (
	"carsawa/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// DealerRepository defines operations for managing dealer accounts.
type DealerRepository interface {
	// CreateDealer creates a new dealer record.
	CreateDealer(dealer *models.Dealer) error

	// UpdateDealer updates specific fields of a dealer document.
	UpdateDealer(id string, updateDoc bson.M) error

	// DeleteDealer removes a dealer record.
	DeleteDealer(id string) error

	// GetDealerByIDWithProjection retrieves a dealer by ID with a projection.
	GetDealerByIDWithProjection(id string, projection bson.M) (*models.Dealer, error)

	// GetDealerByEmail retrieves a dealer by email.
	GetDealerByEmail(email string) (*models.Dealer, error)

	// GetDealerByEmailWithProjection retrieves a dealer by email with a projection.
	GetDealerByEmailWithProjection(email string, projection bson.M) (*models.Dealer, error)

	// GetAllDealersWithProjection retrieves all dealers with the given projection.
	GetAllDealersWithProjection(projection bson.M) ([]models.Dealer, error)

	// GetDealerBySlug retrieves a dealer by its public slug.
	GetDealerBySlug(slug string) (*models.Dealer, error)

	//check dealer availability
	IsDealerAvailable(models.DealerBasicRegistrationData) (bool, error)
}

// newContext creates a context with the given timeout.
func newContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
