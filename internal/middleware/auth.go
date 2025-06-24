package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens for API authentication
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
				"code":  "MISSING_AUTH",
			})
		}

		// Extract token
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}

			// Return secret key (should be from config/environment)
			secretKey := "your-secret-key" // TODO: Move to config
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
				"code":  "INVALID_TOKEN",
			})
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, exists := claims["user_id"]; exists {
				c.Locals("user_id", userID)
			}
			if tenantID, exists := claims["tenant_id"]; exists {
				if tid, ok := tenantID.(float64); ok {
					c.Locals("tenant_id", int64(tid))
				}
			}
		}

		return c.Next()
	}
}

// TenantMiddleware ensures tenant ID is present
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if tenant_id is already set by auth
		if c.Locals("tenant_id") != nil {
			return c.Next()
		}

		// Try to get from header as fallback
		tenantHeader := c.Get("X-Tenant-ID")
		if tenantHeader == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing tenant ID",
				"code":  "MISSING_TENANT",
			})
		}

		// Convert to int64
		tenantID, err := strconv.ParseInt(tenantHeader, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid tenant ID",
				"code":  "INVALID_TENANT",
			})
		}

		c.Locals("tenant_id", tenantID)
		return c.Next()
	}
}

// KafkaAuthMiddleware provides additional security for Kafka endpoints
func KafkaAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Additional security for Kafka endpoints
		apiKey := c.Get("X-Kafka-API-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Kafka API key",
				"code":  "MISSING_KAFKA_KEY",
			})
		}

		// Validate API key (should check against config/database)
		validAPIKey := "your-kafka-api-key" // TODO: Move to config
		if apiKey != validAPIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Kafka API key",
				"code":  "INVALID_KAFKA_KEY",
			})
		}

		return c.Next()
	}
}

// RateLimitMiddleware provides rate limiting (placeholder implementation)
func RateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implement rate limiting based on tenant configuration
		// For now, just pass through
		return c.Next()
	}
}

// LoggingMiddleware adds structured logging
func LoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add request-specific fields to context
		c.Locals("start_time", time.Now())

		err := c.Next()

		// Log request details after processing
		// This could be enhanced with more detailed logging

		return err
	}
}
