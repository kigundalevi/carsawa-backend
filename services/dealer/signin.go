// File: dealer/service.go
package dealer

import (
	"context"
	"fmt"
	"time"

	"carsawa/models"
	"carsawa/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type OTPPendingError struct {
	SessionID string
}

func (e OTPPendingError) Error() string {
	return "OTP verification required for new device"
}

func (s *dealerService) AuthenticateDealer(
	ctx context.Context,
	email string,
	password string,
	currentDevice models.Device,
	providedSessionID string,
) (*models.DealerAuthResponse, error) {
	logger := utils.GetLogger()

	// 1. Fetch dealer with necessary fields
	projection := bson.M{
		"security.password_hash": 1,
		"id":                     1,
		"profile":                1,
		"devices":                1,
		"store":                  1,
		"createdAt":              1,
	}
	dealer, err := s.repo.GetDealerByEmailWithProjection(email, projection)
	if err != nil {
		logger.Error("Failed to fetch dealer", zap.Error(err))
		return nil, fmt.Errorf("authentication failed")
	}
	if dealer == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 2. Verify password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(dealer.Security.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 3. Session management
	sessionClient := utils.GetAuthCacheClient()
	sessionID := providedSessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("dealer:%s:%s", dealer.ID, currentDevice.DeviceID)
		authSession := utils.AuthSession{
			UserID: dealer.ID,
			Email:  dealer.Profile.Contact.Email,
			Device: utils.DeviceSessionInfo{
				DeviceID:   currentDevice.DeviceID,
				DeviceName: currentDevice.DeviceName,
				IP:         currentDevice.IP,
				Location:   currentDevice.Location,
			},
			Status:        "pending",
			CreatedAt:     time.Now(),
			LastUpdatedAt: time.Now(),
		}
		if err := utils.SaveAuthSession(sessionClient, sessionID, authSession); err != nil {
			return nil, fmt.Errorf("failed to start auth session: %w", err)
		}
	}

	// 4. Check existing session
	authSession, err := utils.GetAuthSession(sessionClient, sessionID)
	if err != nil {
		return nil, fmt.Errorf("authentication session expired")
	}

	// 5. Device verification
	deviceExists := false
	for idx, d := range dealer.Devices {
		if d.DeviceID == currentDevice.DeviceID {
			deviceExists = true
			dealer.Devices[idx].IP = currentDevice.IP
			dealer.Devices[idx].Location = currentDevice.Location
			break
		}
	}

	// 6. Handle new devices
	if !deviceExists {
		if authSession.Status != "otp_verified" {
			if len(dealer.Devices) >= 3 {
				return nil, fmt.Errorf("maximum 3 devices allowed")
			}

			// Initiate OTP if not already sent
			otpKey := fmt.Sprintf("otp:%s", sessionID)
			if _, err := sessionClient.Get(ctx, otpKey).Result(); err != nil {
				if err := utils.InitiateDeviceOTP(
					dealer.ID,
					currentDevice.DeviceID,
					dealer.Profile.Contact.Phone,
				); err != nil {
					return nil, fmt.Errorf("failed to send OTP: %w", err)
				}
				authSession.Status = "pending_otp"
				utils.SaveAuthSession(sessionClient, sessionID, *authSession)
			}
			return nil, OTPPendingError{SessionID: sessionID}
		}

		// Add new device after OTP verification
		currentDevice.LastLogin = time.Now()
		currentDevice.Creator = false
		dealer.Devices = append(dealer.Devices, currentDevice)
	}

	// 7. Token generation
	cacheKey := fmt.Sprintf("%s:%s:%s", utils.AuthCachePrefix, dealer.ID, currentDevice.DeviceID)
	if err := sessionClient.Del(ctx, cacheKey).Err(); err != nil {
		logger.Warn("Failed to clear old token cache", zap.Error(err))
	}

	token, err := utils.GenerateToken(dealer.ID, dealer.Profile.Contact.Email, currentDevice.DeviceID)
	if err != nil {
		logger.Error("Token generation failed", zap.Error(err))
		return nil, fmt.Errorf("authentication failed")
	}
	tokenHash := utils.HashToken(token)

	// 8. Update device record
	deviceUpdated := false
	now := time.Now()
	for idx, d := range dealer.Devices {
		if d.DeviceID == currentDevice.DeviceID {
			dealer.Devices[idx].TokenHash = tokenHash
			dealer.Devices[idx].LastLogin = now
			deviceUpdated = true
			break
		}
	}

	if !deviceUpdated {
		currentDevice.TokenHash = tokenHash
		currentDevice.LastLogin = now
		dealer.Devices = append(dealer.Devices, currentDevice)
	}

	// 9. Persist changes
	update := bson.M{
		"$set": bson.M{
			"devices":   dealer.Devices,
			"updatedAt": now,
		},
	}
	if err := s.repo.UpdateDealer(dealer.ID, update); err != nil {
		logger.Error("Failed to update dealer devices", zap.Error(err))
		return nil, fmt.Errorf("authentication failed")
	}

	// 10. Cleanup session
	_ = utils.DeleteAuthSession(sessionClient, sessionID)

	return &models.DealerAuthResponse{
		ID:            dealer.ID,
		Token:         token,
		DealerProfile: dealer.Profile,
		CreatedAt:     dealer.CreatedAt,
	}, nil
}
