package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/platform/observability"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}

		ctx := context.WithValue(
			c.Request.Context(),
			observability.TraceIDKey,
			traceID,
		)

		c.Set(observability.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}
