package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/features/history/service"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
)

type Handler struct {
	log     *slog.Logger
	service *service.Service
}

func NewHandler(log *slog.Logger, service *service.Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	myHistory := router.Group("/my/history")
	{
		myHistory.GET("", h.ListRecent)
		myHistory.GET("/manga/:manga_id", h.ListByManga)
		myHistory.PUT("", h.MarkChaptersRead)
		myHistory.DELETE("", h.UnmarkChaptersRead)
	}
}

func (h *Handler) ListRecent(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var q service.HistoryListQuery
	err := httptransport.BindQuery(ctx, &q, h.log)
	if h.fail(ctx, err) {
		return
	}

	paged, err := h.service.ListRecent(ctx.Request.Context(), ur, q)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, paged)
}

func (h *Handler) ListByManga(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	mangaID, err := h.mangaIDFromPath(ctx)
	if h.fail(ctx, err) {
		return
	}

	var q service.HistoryListQuery
	err = httptransport.BindQuery(ctx, &q, h.log)
	if h.fail(ctx, err) {
		return
	}

	paged, err := h.service.ListByManga(ctx.Request.Context(), ur, mangaID, q)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, paged)
}

func (h *Handler) MarkChaptersRead(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var req service.MarkChaptersAsReadDTO
	err := httptransport.BindJSON(ctx, &req, h.log)
	if h.fail(ctx, err) {
		return
	}

	err = h.service.MarkChaptersRead(ctx.Request.Context(), ur, req)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, nil)
}

func (h *Handler) UnmarkChaptersRead(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var req service.UnmarkChaptersAsReadDTO
	err := httptransport.BindJSON(ctx, &req, h.log)
	if h.fail(ctx, err) {
		return
	}

	err = h.service.UnmarkChaptersRead(ctx.Request.Context(), ur, req)
	if h.fail(ctx, err) {
		return
	}
	httptransport.SuccessResponse(ctx, http.StatusOK, nil)
}
