package handlers

import (
	"carsawa/models"
	"carsawa/services/listing"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ListingHandler struct {
	service listing.ListingService
	logger  *zap.Logger
}

func NewListingHandler(service listing.ListingService, logger *zap.Logger) *ListingHandler {
	return &ListingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ListingHandler) CreateDealerListing(c *gin.Context) {
	var input struct {
		DealerID string         `json:"dealerId"`
		Price    float64        `json:"price"`
		Car      models.Listing `json:"car"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	listing, err := h.service.CreateDealerListing(c.Request.Context(), input.DealerID, input.Car, input.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, listing)
}

func (h *ListingHandler) GetListing(c *gin.Context) {
	id := c.Param("id")
	listing, err := h.service.GetListing(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get listing", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listing)
}

func (h *ListingHandler) UpdateListing(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}

	listing, err := h.service.UpdateListing(c.Request.Context(), id, updates)
	if err != nil {
		h.logger.Error("Update failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listing)
}

func (h *ListingHandler) DeleteListing(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteListing(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete listing", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
