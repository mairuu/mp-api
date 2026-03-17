package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/features/library/service"
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
	library := router.Group("my/library")
	{
		library.GET("", h.GetLibrarySummary)
		library.GET("mangas", h.GetLibrary)
		library.PUT("mangas", h.UpsertLibraryMangas)
		library.GET("mangas/:manga_id", h.GetLibraryManga)
	}
}

func (h *Handler) GetLibrary(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)
	lib, err := h.service.GetLibrary(ctx.Request.Context(), ur)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	httptransport.SuccessResponse(ctx, http.StatusOK, lib)
}

func (h *Handler) GetLibrarySummary(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)
	summary, err := h.service.GetLibrarySummary(ctx.Request.Context(), ur)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	httptransport.SuccessResponse(ctx, http.StatusOK, summary)
}

func (h *Handler) UpsertLibraryMangas(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var req []service.UpsertLibraryMangaDTO
	if err := httptransport.BindJSON(ctx, &req, h.log); err != nil {
		h.handleError(ctx, err)
		return
	}

	if err := h.service.UpsertLibraryMangas(ctx.Request.Context(), ur, req); err != nil {
		h.handleError(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, nil)
}

func (h *Handler) GetLibraryManga(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)
	mangaID, err := h.mangaIDFromPath(ctx)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	manga, err := h.service.GetLibraryManga(ctx.Request.Context(), ur, mangaID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	httptransport.SuccessResponse(ctx, http.StatusOK, manga)
}
