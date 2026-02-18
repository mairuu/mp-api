package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
		}

		lvl := slog.LevelInfo
		msg := "request completed"

		if len(c.Errors) > 0 {
			lvl = slog.LevelError
			msg = "request failed"
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		logger.LogAttrs(c.Request.Context(), lvl, msg, attrs...)
	}
}
