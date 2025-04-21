package handlers

import (
	"net/http"
	"time"

	"carsawa/models"
	"carsawa/services/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	Service user.UserService
	Logger  *zap.Logger
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	// Extract device details
	device := models.Device{
		DeviceID:   c.GetString("deviceID"),
		DeviceName: c.GetString("deviceName"),
		IP:         c.GetString("deviceIP"),
		Location:   c.GetString("deviceLocation"),
		LastLogin:  time.Now(),
	}

	var req models.UserRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	switch req.Step {
	case "basic":
		if req.BasicData == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing basic data"})
			return
		}

		sessionID, code, err := h.Service.InitiateRegistration(c.Request.Context(), *req.BasicData, device)
		if err != nil {
			h.Logger.Error("Registration initiation failed", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"sessionID": sessionID, "status": code})

	case "otp":
		if req.SessionID == "" || req.OTP == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session ID or OTP"})
			return
		}

		code, err := h.Service.VerifyRegistrationOTP(c.Request.Context(), req.SessionID, device.DeviceID, req.OTP)
		if err != nil {
			h.Logger.Error("OTP verification failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"sessionID": req.SessionID, "status": code})

	case "preferences":
		if req.SessionID == "" || len(req.Preferences) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session ID or preferences"})
			return
		}

		authResp, err := h.Service.FinalizeRegistration(c.Request.Context(), req.SessionID, req.Preferences)
		if err != nil {
			h.Logger.Error("Registration finalization failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, authResp)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration step"})
	}
}
