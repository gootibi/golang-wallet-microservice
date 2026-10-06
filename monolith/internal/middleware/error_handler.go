package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	customError "github.com/gootibi/golang-wallet-microservice/monolith/internal/errors"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/logger"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Next handler
		c.Next()

		// Check if any error occured
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// check if the error is one of our custom AppError
			if appErr, ok := err.(*customError.AppError); ok {
				logger.Warn(c.Request.Context(), "Client error occured",
					"code", appErr.Code,
					"message", appErr.Message,
					"status", appErr.StatusCode,
				)
				c.JSON(appErr.StatusCode, gin.H{
					"success": false,
					"error":   appErr,
				})
				return
			}

			// if error is not cover from our custom AppError
			logger.Error(c.Request.Context(), "Unhandled error occured", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   customError.ErrInternalServer,
			})
		}
	}
}
