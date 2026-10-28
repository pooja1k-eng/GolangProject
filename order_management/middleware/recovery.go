package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		ErrorCount.Inc()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error occurred",
		})
	})
}