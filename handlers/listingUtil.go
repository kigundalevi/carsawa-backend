package handlers

import (
	"carsawa/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *ListingHandler) PublishListing(c *gin.Context) {
	listingID := c.Param("id")
	dealerID := c.Query("dealerID")

	listing, err := h.service.PublishListing(c.Request.Context(), listingID, dealerID)
	if err != nil {
		h.logger.Error("Failed to publish listing", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listing)
}

func (h *ListingHandler) CloseListing(c *gin.Context) {
	listingID := c.Param("id")
	isDealer := c.Query("isDealer") == "true"
	ownerID := c.Query("ownerID")

	err := h.service.CloseListing(c.Request.Context(), listingID, ownerID, isDealer)
	if err != nil {
		h.logger.Error("Failed to close listing", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *ListingHandler) SearchListings(c *gin.Context) {
	var filter models.ListingFilter

	// Capture page and limit as query params
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := (page - 1) * limit
	pagination := models.Pagination{Limit: limit, Offset: offset}

	// Optional: extract filters from body (or change this to query if needed)
	if err := c.ShouldBindJSON(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter data"})
		return
	}

	results, err := h.service.SearchListings(c.Request.Context(), filter, pagination)
	if err != nil {
		h.logger.Error("Search error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}
func (h *ListingHandler) GetFeed(c *gin.Context) {
	var filter models.ListingFilter

	// Defaults
	page := 1
	limit := 20

	// Parse query params for page and limit
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := (page - 1) * limit
	pagination := models.Pagination{Limit: limit, Offset: offset}

	// Bind filters from JSON body
	if err := c.ShouldBindJSON(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter"})
		return
	}

	resp, err := h.service.GetFeed(c.Request.Context(), filter, pagination)
	if err != nil {
		h.logger.Error("Failed to fetch feed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
