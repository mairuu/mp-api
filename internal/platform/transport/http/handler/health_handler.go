package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	log *slog.Logger
}

func NewHealthHandler(log *slog.Logger) *HealthHandler {
	return &HealthHandler{log: log}
}

func (h *HealthHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", h.HealthCheck)
}

func (h *HealthHandler) HealthCheck(ctx *gin.Context) {
	h.log.InfoContext(ctx.Request.Context(), "health check request received")
	ctx.String(http.StatusOK, "ok")
}
