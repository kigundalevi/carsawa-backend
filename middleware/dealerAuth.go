package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	dealerRepo "carsawa/database/repository/dealer"
	"carsawa/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// JWTAuthDealerMiddleware authenticates dealers with device validation
func JWTAuthDealerMiddleware(dealerRepo dealerRepo.DealerRepository, optional bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				zap.L().Error("JWTAuthDealerMiddleware: panic recovered", zap.Any("panic", r))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  500,
				})
			}
		}()

		logger := zap.L()
		ctx := context.Background()
		c.Set("isDealerFullAccess", false)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			if optional {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing or invalid Authorization header",
				"code":  0,
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			if optional {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  0,
			})
			return
		}

		// Extract both dealer ID and device ID from token
		dealerID, tokenDeviceID, err := utils.ExtractIDsFromToken(tokenString)
		if err != nil || dealerID == "" || tokenDeviceID == "" {
			if optional {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing dealer/device ID",
				"code":  0,
			})
			return
		}

		// Get device ID from context
		ctxDeviceIDVal, exists := c.Get("deviceID")
		if !exists {
			if optional {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing device context",
				"code":  0,
			})
			return
		}
		ctxDeviceID, ok := ctxDeviceIDVal.(string)
		if !ok || ctxDeviceID == "" {
			if optional {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid device context",
				"code":  0,
			})
			return
		}

		// Validate device match
		if tokenDeviceID != ctxDeviceID {
			logger.Warn("Device ID mismatch",
				zap.String("tokenDevice", tokenDeviceID),
				zap.String("contextDevice", ctxDeviceID),
			)
			if !optional {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Device mismatch",
					"code":  0,
				})
				return
			}
			c.Next()
			return
		}

		tokenHash := utils.HashToken(tokenString)
		if tokenHash == "" {
			if optional {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token hash",
				"code":  0,
			})
			return
		}

		// Use composite cache key with dealerID:deviceID
		cacheKey := utils.AuthCachePrefix + dealerID + ":" + tokenDeviceID
		authCache := utils.GetAuthCacheClient()

		// Cache check logic
		if authCache != nil {
			cachedHash, err := authCache.Get(ctx, cacheKey).Result()
			if err == nil {
				if cachedHash == tokenHash {
					_ = authCache.Expire(ctx, cacheKey, time.Hour).Err()
					c.Set("isDealerFullAccess", true)
					c.Set("dealerID", dealerID)
					c.Next()
					return
				}
				logger.Error("Token hash mismatch in cache", zap.String("dealerID", dealerID))
			} else if err != redis.Nil {
				logger.Error("Cache check failed", zap.Error(err))
			}
		}

		// Database fallback
		proj := bson.M{"id": 1, "security.authTokens": 1}
		dealer, err := dealerRepo.GetDealerByIDWithProjection(dealerID, proj)
		if err != nil || dealer == nil {
			logger.Error("Dealer not found", zap.String("dealerID", dealerID), zap.Error(err))
			if !optional {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Dealer not found",
					"code":  0,
				})
				return
			}
			c.Next()
			return
		}

		if tokenHash != dealer.Security.TokenHash {
			logger.Error("Token hash mismatch in DB", zap.String("dealerID", dealerID))
			if !optional {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Token mismatch",
					"code":  0,
				})
				return
			}
			c.Next()
			return
		}

		// Update cache
		if authCache != nil {
			if err := authCache.Set(ctx, cacheKey, tokenHash, time.Hour).Err(); err != nil {
				logger.Error("Failed to update cache", zap.Error(err))
			}
		}

		c.Set("isDealerFullAccess", true)
		c.Set("dealerID", dealerID)
		c.Next()
	}
}
