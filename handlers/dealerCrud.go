package handlers

import (
	"net/http"

	"carsawa/services/dealer"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type DealerHandler struct {
	service dealer.DealerService
	logger  *zap.Logger
}

func NewDealerHandler(s dealer.DealerService, logger *zap.Logger) *DealerHandler {
	return &DealerHandler{service: s, logger: logger}
}

func (h *DealerHandler) GetDealer(c *gin.Context) {
	id := c.Param("id")

	// Pass the request context containing the access level
	dealer, err := h.service.GetDealer(c.Request.Context(), id)
	if err != nil {
		h.logger.Warn("get dealer failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Dealer not found"})
		return
	}

	// Sanitize output based on access level
	if !c.GetBool("isDealerFullAccess") {
		c.JSON(http.StatusOK, gin.H{
			"id":      dealer.ID,
			"profile": dealer.Profile,
			"store":   dealer.Store,
		})
		return
	}
	c.JSON(http.StatusOK, dealer)
}

func (h *DealerHandler) UpdateDealer(c *gin.Context) {
	id := c.Param("id")
	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.UpdateDealer(c.Request.Context(), id, updates)
	if err != nil {
		h.logger.Error("update failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *DealerHandler) DeleteDealer(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteDealer(c.Request.Context(), id); err != nil {
		h.logger.Error("delete failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DealerHandler) ListDealers(c *gin.Context) {
	dealers, err := h.service.ListDealers(c.Request.Context())
	if err != nil {
		h.logger.Error("list failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dealers)
}
