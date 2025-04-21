package dealer

import (
	"carsawa/models"
	"carsawa/utils"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// Helper Functions
func generateSlug(name string) string {
	slug := strings.ToLower(strings.ReplaceAll(
		strings.TrimSpace(name),
		" ", "-",
	))
	// Remove special characters
	return regexp.MustCompile(`[^\w-]`).ReplaceAllString(slug, "")
}

func ensureUniqueSlug(s *dealerService, baseSlug string) (string, error) {
	slug := baseSlug
	for i := 1; ; i++ {
		_, err := s.repo.GetDealerBySlug(slug)
		if err != nil {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
		if i > 10 {
			return "", fmt.Errorf("failed to generate unique slug after 10 attempts")
		}
	}
}

// Auth Management
func (s *dealerService) RevokeDealerAuthToken(dealerID, deviceID string) error {
	dealer, err := s.repo.GetDealerByIDWithProjection(dealerID, nil)
	if err != nil {
		return fmt.Errorf("dealer lookup failed: %w", err)
	}

	deviceFound := false
	for i, d := range dealer.Devices {
		if d.DeviceID == deviceID {
			dealer.Devices[i].TokenHash = ""
			deviceFound = true
			break
		}
	}

	if !deviceFound {
		return fmt.Errorf("device not registered")
	}

	updateDoc := bson.M{
		"$set": bson.M{
			"devices":   dealer.Devices,
			"updatedAt": time.Now(),
		},
	}

	if err := s.repo.UpdateDealer(dealerID, updateDoc); err != nil {
		return fmt.Errorf("database update failed: %w", err)
	}

	// Clear Redis cache
	cacheKey := fmt.Sprintf("%s:%s", dealerID, deviceID)
	authCache := utils.GetAuthCacheClient()
	if err := authCache.Del(context.Background(), cacheKey).Err(); err != nil {
		utils.GetLogger().Warn("Cache cleanup failed",
			zap.String("key", cacheKey),
			zap.Error(err),
		)
	}

	return nil
}

// Validation
func validateBasicRegistrationData(basicReq models.DealerBasicRegistrationData) error {
	// Dealer Name validation
	if strings.TrimSpace(basicReq.DealerName) == "" {
		return fmt.Errorf("dealer name required")
	}
	if len(basicReq.DealerName) > 100 {
		return fmt.Errorf("dealer name too long (max 100 chars)")
	}

	// Email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(basicReq.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Password validation
	if len(basicReq.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(basicReq.Password) {
		return fmt.Errorf("password must contain uppercase letter")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(basicReq.Password) {
		return fmt.Errorf("password must contain number")
	}

	// Phone validation
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{7,14}$`)
	if !phoneRegex.MatchString(basicReq.PhoneNumber) {
		return fmt.Errorf("invalid phone number format")
	}

	// Address validation
	if len(basicReq.Address) < 10 {
		return fmt.Errorf("address too short")
	}

	// Geo validation
	if basicReq.LocationGeo.Type != "Point" || len(basicReq.LocationGeo.Coordinates) != 2 {
		return fmt.Errorf("invalid geolocation format")
	}
	lat, lng := basicReq.LocationGeo.Coordinates[1], basicReq.LocationGeo.Coordinates[0]
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return fmt.Errorf("invalid geolocation coordinates")
	}

	return nil
}

func GenerateDealerID() string {
	return uuid.New().String()
}

func buildAuthResponse(dealer *models.Dealer, token string) *models.DealerAuthResponse {

	return &models.DealerAuthResponse{
		ID:            dealer.ID,
		Token:         token,
		DealerProfile: dealer.Profile,
		CreatedAt:     dealer.CreatedAt,
	}
}

// Helper functions
func createDealerFromSession(session models.DealerRegistrationSession) (*models.Dealer, error) {
	return &models.Dealer{
		ID: GenerateDealerID(),
		Profile: models.DealerProfile{
			DealerName: session.BasicData.DealerName,
			Slug:       generateSlug(session.BasicData.DealerName),
			Contact: models.Contact{
				Email:    session.BasicData.Email,
				Phone:    session.BasicData.PhoneNumber,
				WhatsApp: session.BasicData.PhoneNumber,
			},
			Location: models.Location{
				Address:  session.BasicData.Address,
				GeoPoint: session.BasicData.LocationGeo,
				City:     session.BasicData.City,
			},
			Rating: 0,
		},
		Store: models.Store{
			ServiceCatalog: session.ServiceCatalogue,
			Branding: models.StoreBranding{
				PrimaryColor:   "#2563eb", // Default brand color
				SecondaryColor: "#1d4ed8",
			},
		},
		Verification: models.Verification{
			Level:      "basic",
			Status:     session.VerificationStatus,
			Documents:  []string{session.KYPData.DocumentURL, session.KYPData.SelfieURL},
			VerifiedAt: time.Now(),
		},
		Security:  models.Security{},
		Devices:   session.Devices,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func generateAuthToken(dealerID, email, deviceID string) (string, string, error) {
	token, err := utils.GenerateToken(dealerID, email, deviceID)
	if err != nil {
		return "", "", fmt.Errorf("token generation failed: %w", err)
	}
	return token, utils.HashToken(token), nil
}

func updateDeviceToken(devices []models.Device, deviceID, deviceName, tokenHash string) []models.Device {
	now := time.Now()
	for i := range devices {
		if devices[i].DeviceID == deviceID {
			devices[i].TokenHash = tokenHash
			devices[i].LastLogin = now
			return devices
		}
	}
	return append(devices, models.Device{
		DeviceID:   deviceID,
		TokenHash:  tokenHash,
		LastLogin:  now,
		DeviceName: deviceName,
	})
}
