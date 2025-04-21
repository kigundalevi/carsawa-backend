package dealer

import (
	"context"

	dealerRepo "carsawa/database/repository/dealer"
	listingRepo "carsawa/database/repository/listing"
	"carsawa/models"
	"carsawa/services/notification"
	"carsawa/utils/email"
	"carsawa/utils/token"

	"go.mongodb.org/mongo-driver/bson"
)

type DealerService interface {
	// Core dealer operations
	UpdateDealer(ctx context.Context, id string, updates bson.M) (*models.Dealer, error)
	GetDealer(ctx context.Context, id string) (*models.Dealer, error)
	GetDealerByEmail(ctx context.Context, email string) (*models.Dealer, error)
	ListDealers(ctx context.Context) ([]models.Dealer, error)
	DeleteDealer(ctx context.Context, id string) error

	// Registration flow methods
	RegisterBasic(
		basicReq models.DealerBasicRegistrationData,
		device models.Device,
	) (string, int, error)

	VerifyOTP(
		sessionID string,
		deviceID string,
		providedOTP string,
	) (int, error)

	VerifyKYP(
		sessionID string,
		kypData models.KYPVerificationData,
	) (int, error)

	FinalizeRegistration(
		sessionID string,
		catalogueData models.ServiceCatalogue,
	) (*models.DealerAuthResponse, error)

	// Auth & security
	AuthenticateDealer(
		ctx context.Context,
		email string,
		password string,
		currentDevice models.Device,
		providedSessionID string,
	) (*models.DealerAuthResponse, error)
}

type dealerService struct {
	repo          dealerRepo.DealerRepository
	listingsRepo  listingRepo.ListingRepository
	tokenProvider token.Provider
	emailService  email.EmailService
	notifier      notification.NotificationService
}

func NewDealerService(
	repo dealerRepo.DealerRepository,
	listingsRepo listingRepo.ListingRepository,
	tp token.Provider,
	es email.EmailService,
	not notification.NotificationService,
) DealerService {
	return &dealerService{
		repo:          repo,
		listingsRepo:  listingsRepo,
		tokenProvider: tp,
		emailService:  es,
		notifier:      not,
	}
}
