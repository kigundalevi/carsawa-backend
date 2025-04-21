package handlers

import (
	dealerRepo "carsawa/database/repository/dealer"
	userRepo "carsawa/database/repository/user"

	"github.com/gin-gonic/gin"
)

type HandlerBundle struct {
	// Repositories
	DealerRepo dealerRepo.DealerRepository
	UserRepo   userRepo.UserRepository

	// Services
	ListingService listing
	AuthService    services.AuthService

	// Dealer Handlers
	RegisterDealerHandler      func(c *gin.Context)
	LoginDealerHandler         func(c *gin.Context)
	LogoutDealerHandler        func(c *gin.Context)
	GetDealerProfileHandler    func(c *gin.Context)
	UpdateDealerProfileHandler func(c *gin.Context)
	CreateListingHandler       func(c *gin.Context)
	UpdateListingHandler       func(c *gin.Context)
	DeleteListingHandler       func(c *gin.Context)
	GetDealerListingsHandler   func(c *gin.Context)
	GetTradeInLeadsHandler     func(c *gin.Context)
	ContactUserHandler         func(c *gin.Context)
	PlaceBidOnUserCarHandler   func(c *gin.Context)

	// User Handlers
	RegisterUserHandler               func(c *gin.Context)
	LoginUserHandler                  func(c *gin.Context)
	LogoutUserHandler                 func(c *gin.Context)
	GetCurrentUserHandler             func(c *gin.Context)
	UpdateUserHandler                 func(c *gin.Context)
	DeleteUserHandler                 func(c *gin.Context)
	RevokeUserAuthTokenHandler        func(c *gin.Context)
	UpdateUserPasswordHandler         func(c *gin.Context)
	CreateMyCarHandler                func(c *gin.Context)
	GetMyCarsHandler                  func(c *gin.Context)
	GetMyCarByIDHandler               func(c *gin.Context)
	DeleteMyCarHandler                func(c *gin.Context)
	GetMyCarBidsHandler               func(c *gin.Context)
	AcceptDealerBidHandler            func(c *gin.Context)
	CreateTradeInHandler              func(c *gin.Context)
	GetUserTradeInsHandler            func(c *gin.Context)
	DeleteTradeInHandler              func(c *gin.Context)
	GetTradeInOffersHandler           func(c *gin.Context)
	GetNotificationsHandler           func(c *gin.Context)
	MarkNotificationsReadHandler      func(c *gin.Context)
	MarkAllNotificationsReadHandler   func(c *gin.Context)
	GetUnreadNotificationCountHandler func(c *gin.Context)
	GetPublicTradeInsHandler          func(c *gin.Context)

	// Public/Feed Handlers
	GetListingsHandler         func(c *gin.Context)
	SearchHandler              func(c *gin.Context)
	PublicDealerProfileHandler func(c *gin.Context)

	// Miscellaneous
	UploadFileHandler       func(c *gin.Context)
	GetDownloadURLHandler   func(c *gin.Context)
	GetUserPurchasesHandler func(c *gin.Context)
	GetUserSalesHandler     func(c *gin.Context)
	RecordPurchaseHandler   func(c *gin.Context)
	RecordSaleHandler       func(c *gin.Context)
}

func NewHandlerBundle(
	dealerRepo repositories.DealerRepository,
	userRepo repositories.UserRepository,
	listingService services.ListingService,
	tradeInService services.TradeInService,
	authService services.AuthService,
) *HandlerBundle {
	return &HandlerBundle{
		DealerRepo:     dealerRepo,
		UserRepo:       userRepo,
		ListingService: listingService,
		TradeInService: tradeInService,
		AuthService:    authService,
	}
}
