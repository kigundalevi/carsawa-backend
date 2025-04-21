package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"carsawa/config"
	"carsawa/database"

	"carsawa/handlers"
	"carsawa/middleware"
	"carsawa/routes"
	"carsawa/utils"
	"carsawa/utils/email"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()
	logger := utils.GetLogger()

	database.InitDB()
	utils.InitRedis()

	storageService, err := utils.Cloudinary()
	if err != nil {
		logger.Sugar().Fatalf("failed to init storage: %v", err)
	}

	router := gin.New()
	router.Use(
		gin.Recovery(),
		utils.ErrorHandler(),
		gin.Logger(),
		middleware.RateLimitMiddleware(),
		middleware.GeolocationMiddleware(),
	)

	emailSvc := email.NewSMTPEmailService(smtpCfg)

	userSvc := user.NewUserService(userRepo, jwtProvider, emailSvc)
	dealerSvc := dealer.NewDealerService(dealerRepo, jwtProvider, emailSvc)

	userRepo := user.NewMongoUserRepo()
	dealerRepo := dealer.NewMongoDealerRepo()
	carRepo := car.NewMongoCarRepo()
	transactionRepo := transaction.NewMongoTransactionRepo()

	userHandler := handlers.NewUserHandler(userRepo)
	dealerHandler := handlers.NewDealerHandler(dealerRepo)
	carHandler := handlers.NewCarHandler(carRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionRepo, carRepo)
	storageHandler := handlers.NewStorageHandler(storageService)

	hb := &handlers.HandlerBundle{
		UserRepo:   userRepo,
		DealerRepo: dealerRepo,

		RegisterUserHandler:        userHandler.RegisterUser,
		LoginUserHandler:           userHandler.LoginUser,
		LogoutUserHandler:          userHandler.LogoutUser,
		GetCurrentUserHandler:      userHandler.GetCurrentUser,
		GetUserByIDHandler:         userHandler.GetUserByID,
		GetUserByEmailHandler:      userHandler.GetUserByEmail,
		UpdateUserHandler:          userHandler.UpdateUser,
		DeleteUserHandler:          userHandler.DeleteUser,
		RevokeUserAuthTokenHandler: userHandler.RevokeAuthToken,
		UpdateUserPasswordHandler:  userHandler.UpdatePassword,

		RegisterDealerHandler:        dealerHandler.RegisterDealer,
		LoginDealerHandler:           dealerHandler.LoginDealer,
		LogoutDealerHandler:          dealerHandler.LogoutDealer,
		GetCurrentDealerHandler:      dealerHandler.GetCurrentDealer,
		GetDealerByIDHandler:         dealerHandler.GetDealerByID,
		UpdateDealerHandler:          dealerHandler.UpdateDealer,
		DeleteDealerHandler:          dealerHandler.DeleteDealer,
		RevokeDealerAuthTokenHandler: dealerHandler.RevokeAuthToken,
		UpdateDealerPasswordHandler:  dealerHandler.UpdatePassword,

		GetAllCarsHandler:      carHandler.GetAllCars,
		GetCarByIDHandler:      carHandler.GetCarByID,
		CreateCarHandler:       carHandler.CreateCar,
		UpdateCarHandler:       carHandler.UpdateCar,
		DeleteCarHandler:       carHandler.DeleteCar,
		UpdateCarStatusHandler: carHandler.UpdateCarStatus,

		CreateMyCarHandler:       userHandler.CreateMyCar,
		GetMyCarsHandler:         userHandler.GetMyCars,
		GetMyCarByIDHandler:      userHandler.GetMyCarByID,
		DeleteMyCarHandler:       userHandler.DeleteMyCar,
		GetMyCarBidsHandler:      userHandler.GetMyCarBids,
		AcceptDealerBidHandler:   userHandler.AcceptDealerBid,
		PlaceBidOnUserCarHandler: dealerHandler.PlaceBidOnUserCar,

		GetUserPurchasesHandler: transactionHandler.GetUserPurchases,
		GetUserSalesHandler:     transactionHandler.GetUserSales,
		RecordPurchaseHandler:   transactionHandler.RecordPurchase,
		RecordSaleHandler:       transactionHandler.RecordSale,

		GetNotificationsHandler:           userHandler.GetNotifications,
		MarkNotificationsReadHandler:      userHandler.MarkNotificationsRead,
		MarkAllNotificationsReadHandler:   userHandler.MarkAllNotificationsRead,
		GetUnreadNotificationCountHandler: userHandler.GetUnreadNotificationCount,

		UploadFileHandler:     storageHandler.UploadFile,
		GetDownloadURLHandler: storageHandler.GetDownloadURL,
	}

	routes.RegisterRoutes(router, hb)

	port := config.AppConfig.AppPort
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: router,
	}

	logger.Sugar().Infof("Server starting on %s...", srv.Addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Sugar().Fatalf("server start error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Sugar().Fatalf("server forced shutdown: %v", err)
	}

	logger.Sugar().Info("server stopped gracefully")
}
