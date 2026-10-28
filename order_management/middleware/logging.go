package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		reqID, _ := c.Get(RequestIDHeader)
		errStr := c.Errors.ByType(gin.ErrorTypePrivate).String()

		logLine := fmt.Sprintf(
			"request_id=%v method=%s path=%s status=%d duration=%v error=%s",
			reqID, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration, errStr,
		)
		fmt.Println(logLine)
	}
}