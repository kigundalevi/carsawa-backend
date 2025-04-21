// File: dealer/service.go
package dealer

import (
	"carsawa/models"
	"carsawa/utils"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *dealerService) RegisterBasic(basicReq models.DealerBasicRegistrationData, device models.Device) (string, int, error) {
	// Sanitize input
	basicReq.Email = strings.ToLower(strings.TrimSpace(basicReq.Email))
	basicReq.PhoneNumber = utils.NormalizePhoneNumber(basicReq.PhoneNumber)

	if err := validateBasicRegistrationData(basicReq); err != nil {
		return "", 0, fmt.Errorf("validation error: %w", err)
	}

	available, err := s.repo.IsDealerAvailable(basicReq)
	if err != nil {
		return "", 0, fmt.Errorf("availability check failed: %w", err)
	}
	if available {
		return "", 0, fmt.Errorf("a dealer with this email or phone already exists")
	}

	sessionID := utils.GenerateSessionID()
	now := time.Now()

	if err := utils.InitiateDeviceOTP(sessionID, device.DeviceID, basicReq.PhoneNumber); err != nil {
		return "", 0, fmt.Errorf("failed to initiate OTP: %w", err)
	}

	session := models.DealerRegistrationSession{
		TempID:        sessionID,
		BasicData:     basicReq,
		OTPStatus:     "pending",
		CreatedAt:     now,
		LastUpdatedAt: now,
		Devices:       []models.Device{device},
	}

	authCacheClient := utils.GetAuthCacheClient()
	if err := utils.SaveRegistrationSession(authCacheClient, sessionID, session, 30*time.Minute); err != nil {
		return "", 0, fmt.Errorf("failed to save registration session: %w", err)
	}

	return sessionID, 100, nil
}

func (s *dealerService) VerifyOTP(sessionID string, deviceID string, providedOTP string) (int, error) {
	authCacheClient := utils.GetAuthCacheClient()

	session, err := utils.GetRegistrationSession(authCacheClient, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve registration session: %w", err)
	}

	if err := utils.VerifyDeviceOTPRecord(sessionID, deviceID, providedOTP); err != nil {
		return 0, fmt.Errorf("OTP verification failed: %w", err)
	}

	session.OTPStatus = "verified"
	session.LastUpdatedAt = time.Now()
	if err := utils.SaveRegistrationSession(authCacheClient, sessionID, session, 30*time.Minute); err != nil {
		return 0, fmt.Errorf("failed to update OTP status: %w", err)
	}

	return 105, nil
}

func (s *dealerService) VerifyKYP(sessionID string, kypData models.KYPVerificationData) (int, error) {
	authCacheClient := utils.GetAuthCacheClient()

	session, err := utils.GetRegistrationSession(authCacheClient, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve session: %w", err)
	}

	if session.OTPStatus != "verified" {
		return 0, fmt.Errorf("OTP verification required first")
	}

	if kypData.DocumentURL == "" || kypData.LegalName == "" || kypData.SelfieURL == "" {
		return 0, fmt.Errorf("all KYP documents are required")
	}

	session.KYPData = kypData
	session.VerificationStatus = "verified"
	session.VerificationLevel = "basic"
	session.LastUpdatedAt = time.Now()

	if err := utils.SaveRegistrationSession(authCacheClient, sessionID, session, 30*time.Minute); err != nil {
		return 0, fmt.Errorf("failed to update session: %w", err)
	}

	return 101, nil
}

func (s *dealerService) FinalizeRegistration(sessionID string, catalogueData models.ServiceCatalogue) (*models.DealerAuthResponse, error) {
	authCacheClient := utils.GetAuthCacheClient()

	session, err := utils.GetRegistrationSession(authCacheClient, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	if session.VerificationStatus != "verified" {
		return nil, fmt.Errorf("KYP verification required")
	}

	if len(catalogueData.Services) == 0 {
		return nil, fmt.Errorf("at least one service is required")
	}

	session.ServiceCatalogue = catalogueData
	session.LastUpdatedAt = time.Now()

	dealer, err := createDealerFromSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create dealer: %w", err)
	}

	// Password handling
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(session.BasicData.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}
	dealer.Security.PasswordHash = string(hashedPassword)

	// Token generation
	registrationDevice := session.Devices[0]
	token, tokenHash, err := generateAuthToken(dealer.ID, dealer.Profile.Contact.Email, registrationDevice.DeviceID)
	if err != nil {
		return nil, err
	}

	// Update devices
	dealer.Devices = updateDeviceToken(dealer.Devices, registrationDevice.DeviceID, registrationDevice.DeviceName, tokenHash)

	// Persist to database
	if err := s.repo.CreateDealer(dealer); err != nil {
		return nil, fmt.Errorf("database creation failed: %w", err)
	}

	// Cleanup
	if err := utils.DeleteRegistrationSession(authCacheClient, sessionID); err != nil {
		utils.GetLogger().Error("Session cleanup failed",
			zap.String("sessionID", sessionID),
			zap.Error(err),
		)
	}

	return buildAuthResponse(dealer, token), nil
}
