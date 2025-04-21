package routes

import (
	"net/http"
	"time"

	"carsawa/handlers"
	"carsawa/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterDealerRoutes(r *gin.Engine, hb *handlers.HandlerBundle) {
	dealers := r.Group("/api/dealers")
	dealers.Use(middleware.DeviceDetailsMiddleware())
	{
		dealers.POST("/register", hb.RegisterDealerHandler)
		dealers.POST("/login", hb.LoginDealerHandler)
		dealers.POST("/logout", middleware.JWTAuthDealerMiddleware(hb.DealerRepo), hb.LogoutDealerHandler)

		protected := dealers.Group("")
		protected.Use(middleware.JWTAuthDealerMiddleware(hb.DealerRepo))
		{
			protected.GET("/profile", hb.GetDealerProfileHandler)
			protected.PUT("/profile", hb.UpdateDealerProfileHandler)

			protected.POST("/listings", hb.CreateListingHandler)
			protected.PUT("/listings/:id", hb.UpdateListingHandler)
			protected.DELETE("/listings/:id", hb.DeleteListingHandler)
			protected.GET("/listings", hb.GetDealerListingsHandler)

			protected.GET("/trade-ins/leads", hb.GetTradeInLeadsHandler)
			protected.POST("/trade-ins/:id/contact", hb.ContactUserHandler)
		}
	}

	r.GET("/api/dealers/:slug", hb.PublicDealerProfileHandler)
}

func RegisterUserRoutes(r *gin.Engine, hb *handlers.HandlerBundle) {
	users := r.Group("/api/users")
	users.Use(middleware.DeviceDetailsMiddleware())
	{
		users.POST("/register", hb.RegisterUserHandler)
		users.POST("/login", hb.LoginUserHandler)
		users.POST("/logout", middleware.JWTAuthUserMiddleware(hb.UserRepo), hb.LogoutUserHandler)

		protected := users.Group("")
		protected.Use(middleware.JWTAuthUserMiddleware(hb.UserRepo))
		{
			protected.POST("/trade-ins", hb.CreateTradeInHandler)
			protected.GET("/trade-ins", hb.GetUserTradeInsHandler)
			protected.DELETE("/trade-ins/:id", hb.DeleteTradeInHandler)
			protected.GET("/trade-ins/:id/offers", hb.GetTradeInOffersHandler)
		}
	}
}

func RegisterPublicRoutes(r *gin.Engine, hb *handlers.HandlerBundle) {
	r.GET("/api/listings", hb.GetListingsHandler)
	r.GET("/api/trade-ins", hb.GetPublicTradeInsHandler)
	r.GET("/api/search", hb.SearchHandler)
}

func RegisterRoutes(r *gin.Engine, hb *handlers.HandlerBundle) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	RegisterDealerRoutes(r, hb)
	RegisterUserRoutes(r, hb)
	RegisterPublicRoutes(r, hb)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
