package utils

import (
	"carsawa/models"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ID Generation
func GenerateSessionID() string {
	return uuid.New().String()
}

// Session Management
func SaveRegistrationSession(client *redis.Client, sessionID string, session models.DealerRegistrationSession, ttl time.Duration) error {
	ctx := context.Background()
	data, err := json.Marshal(session)
	if err != nil {
		GetLogger().Error("Failed to marshal session", zap.Error(err))
		return fmt.Errorf("session serialization failed")
	}

	if err := client.Set(ctx, sessionID, data, ttl).Err(); err != nil {
		GetLogger().Error("Redis save failed",
			zap.String("sessionID", sessionID),
			zap.Error(err),
		)
		return fmt.Errorf("session storage error")
	}
	return nil
}

func GetRegistrationSession(client *redis.Client, sessionID string) (models.DealerRegistrationSession, error) {
	var session models.DealerRegistrationSession
	ctx := context.Background()

	data, err := client.Get(ctx, sessionID).Result()
	if err != nil {
		GetLogger().Warn("Session not found",
			zap.String("sessionID", sessionID),
			zap.Error(err),
		)
		return session, fmt.Errorf("session expired or invalid")
	}

	if err := json.Unmarshal([]byte(data), &session); err != nil {
		GetLogger().Error("Session deserialization failed",
			zap.String("sessionID", sessionID),
			zap.Error(err),
		)
		return session, fmt.Errorf("invalid session data")
	}
	return session, nil
}

func DeleteRegistrationSession(client *redis.Client, sessionID string) error {
	ctx := context.Background()
	if err := client.Del(ctx, sessionID).Err(); err != nil {
		GetLogger().Error("Session deletion failed",
			zap.String("sessionID", sessionID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to clear session")
	}
	return nil
}

func NormalizePhoneNumber(phone string) string {
	// Remove all non-digit characters
	re := regexp.MustCompile(`\D`)
	cleaned := re.ReplaceAllString(phone, "")

	// Handle Kenyan phone numbers specifically
	if strings.HasPrefix(cleaned, "0") && len(cleaned) == 10 {
		return "+254" + cleaned[1:]
	}

	// Add country code if missing
	if !strings.HasPrefix(cleaned, "+") {
		// Default to Kenya (+254) if no country code
		if len(cleaned) == 9 {
			return "+254" + cleaned
		}
	}

	return "+" + cleaned
}
