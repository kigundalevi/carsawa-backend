package handlers

import (
	"carsawa/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *ListingHandler) CreateUserBidListing(c *gin.Context) {
	var payload struct {
		UserID string         `json:"userId"`
		Car    models.Listing `json:"car"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	listing, err := h.service.CreateUserBidListing(c.Request.Context(), payload.UserID, payload.Car)
	if err != nil {
		h.logger.Error("Failed to create user bid listing", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, listing)
}

func (h *ListingHandler) AddBid(c *gin.Context) {
	listingID := c.Param("id")
	var bid models.Bid
	if err := c.ShouldBindJSON(&bid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bid data"})
		return
	}

	listing, err := h.service.AddBid(c.Request.Context(), listingID, bid)
	if err != nil {
		h.logger.Error("Failed to add bid", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listing)
}

func (h *ListingHandler) AcceptBid(c *gin.Context) {
	listingID := c.Param("id")
	bidID := c.Param("bidID")
	userID := c.Query("userID")

	listing, err := h.service.AcceptBid(c.Request.Context(), listingID, bidID, userID)
	if err != nil {
		h.logger.Error("Failed to accept bid", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listing)
}
