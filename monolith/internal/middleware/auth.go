package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/auth"
	customError "github.com/gootibi/golang-wallet-microservice/monolith/internal/errors"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		autHeader := c.GetHeader("Authorization")
		if autHeader == "" {
			c.Error(customError.NewAppError(http.StatusUnauthorized, "MISSING_TOKEN", "Token is missing"))
			c.Abort()
			return
		}

		// Split the Bearer token
		parts := strings.Split(autHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(customError.NewAppError(http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid, should be Bearer <token>"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.Error(customError.NewAppError(http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired"))
			c.Abort()
			return
		}

		// Save to context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}
