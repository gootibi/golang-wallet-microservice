package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/logger"
)

func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// read from request header
		corID := c.GetHeader("X-Correlation-ID")
		if corID == "" {
			// Generate new uuid
			corID = uuid.New().String()
		}

		// Set in request context
		ctx := context.WithValue(c.Request.Context(), logger.CorrelationIDKey, corID)
		c.Request = c.Request.WithContext(ctx)

		// Set response header
		c.Header("X-Correlation-ID", corID)

		// Continue
		c.Next()
	}
}
