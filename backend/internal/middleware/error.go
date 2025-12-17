package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/pkg/errors"
)

// ErrorHandlerMiddleware handles panics and errors globally
// Reference: BACKEND-IMPLEMENTATION.md lines 1347-1524 (Error Handling)
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic (in production, send to error tracking service)
				// logger.Error("Panic recovered", "error", err, "path", c.Request.URL.Path)

				// Return internal server error
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "An unexpected error occurred",
					},
				})
				c.Abort()
			}
		}()

		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last().Err

			// Check if it's our custom error
			if appErr, ok := err.(*errors.AppError); ok {
				c.JSON(appErr.StatusCode, gin.H{
					"success": false,
					"error":   appErr,
				})
				return
			}

			// Unknown error - return internal server error
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": err.Error(),
				},
			})
		}
	}
}
