package handlers

import (
	"net/http"
	"time"

	"carsawa/models"
	"carsawa/services/dealer"
	"carsawa/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterDealerHandler handles multi-step dealer registration
func (h *DealerHandler) RegisterDealerHandler(c *gin.Context) {
	logger := utils.GetLogger()

	// Extract device details from context
	deviceID, exists := c.Get("deviceID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing device ID"})
		return
	}
	deviceName, exists := c.Get("deviceName")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing device name"})
		return
	}
	deviceIP, _ := c.Get("deviceIP")
	deviceLocation, _ := c.Get("deviceLocation")

	device := models.Device{
		DeviceID:   deviceID.(string),
		DeviceName: deviceName.(string),
		IP:         deviceIP.(string),
		Location:   deviceLocation.(string),
		LastLogin:  time.Now(),
		Creator:    true,
	}

	var req models.DealerRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	switch req.Step {
	case "basic":
		// Step 1: Basic registration + OTP initiation
		if req.BasicData == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing basic registration data"})
			return
		}

		sessionID, status, err := h.service.RegisterBasic(*req.BasicData, device)
		if err != nil {
			logger.Error("Basic registration failed",
				zap.Error(err),
				zap.String("step", "basic"),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "Basic registration failed",
				"detail": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sessionID": sessionID,
			"status":    status,
			"nextStep":  "otp_verification",
		})

	case "otp":
		// Step 1.5: OTP verification
		if req.SessionID == "" || req.OTP == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Session ID and OTP required",
			})
			return
		}

		status, err := h.service.VerifyOTP(req.SessionID, device.DeviceID, req.OTP)
		if err != nil {
			logger.Error("OTP verification failed",
				zap.String("sessionID", req.SessionID),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "OTP verification failed",
				"detail": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sessionID": req.SessionID,
			"status":    status,
			"nextStep":  "kyp_verification",
		})

	case "kyp":
		// Step 2: KYP verification
		if req.SessionID == "" || req.KYPData == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Session ID and KYP data required",
			})
			return
		}

		status, err := h.service.VerifyKYP(req.SessionID, *req.KYPData)
		if err != nil {
			logger.Error("KYP verification failed",
				zap.String("sessionID", req.SessionID),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "KYP verification failed",
				"detail": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sessionID": req.SessionID,
			"status":    status,
			"nextStep":  "service_catalogue",
		})

	case "catalogue":
		// Step 3: Service catalogue & finalization
		if req.SessionID == "" || req.ServiceCatalogue == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Session ID and service catalogue required",
			})
			return
		}

		authResp, err := h.service.FinalizeRegistration(req.SessionID, *req.ServiceCatalogue)
		if err != nil {
			logger.Error("Registration finalization failed",
				zap.String("sessionID", req.SessionID),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Registration completion failed",
				"detail": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, authResp)

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid registration step",
			"valid_steps": []string{
				"basic", "otp", "kyp", "catalogue",
			},
		})
	}
}

func (h *DealerHandler) AuthenticateDealerHandler(c *gin.Context) {
	logger := utils.GetLogger()

	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required"`
		SessionID string `json:"sessionID"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid auth request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get device details from context
	device := models.Device{
		DeviceID:   c.GetString("deviceID"),
		DeviceName: c.GetString("deviceName"),
		IP:         c.GetString("deviceIP"),
		Location:   c.GetString("deviceLocation"),
		LastLogin:  time.Now(),
	}

	authResp, err := h.service.AuthenticateDealer(
		c.Request.Context(),
		req.Email,
		req.Password,
		device,
		req.SessionID,
	)

	if err != nil {
		if otpErr, ok := err.(dealer.OTPPendingError); ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":     otpErr.Error(),
				"sessionID": otpErr.SessionID,
				"nextStep":  "otp_verification",
			})
			return
		}
		logger.Error("Dealer auth failed", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResp)
}
