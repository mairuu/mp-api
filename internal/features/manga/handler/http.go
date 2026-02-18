package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/features/manga/service"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
)

type Handler struct {
	log     *slog.Logger
	service *service.Service
}

func NewHandler(logger *slog.Logger, service *service.Service) *Handler {
	return &Handler{
		log:     logger,
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	mangas := router.Group("mangas")
	{
		mangas.POST("", h.CreateManga)
		mangas.GET("", h.ListMangas)
		mangas.GET(":manga_id", h.GetMangaByID)
		mangas.PUT(":manga_id", h.UpdateManga)
		mangas.DELETE(":manga_id", h.DeleteManga)

	}
}

func (h *Handler) CreateManga(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var req service.CreateMangaDTO
	if err := httptransport.BindJSON(ctx, &req, h.log); err != nil {
		h.handleError(ctx, err)
		return
	}

	m, err := h.service.CreateManga(ctx.Request.Context(), ur, req)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusCreated, m)
}

func (h *Handler) ListMangas(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var pq service.MangaListQuery
	if err := httptransport.BindQuery(ctx, &pq, h.log); err != nil {
		h.handleError(ctx, err)
		return
	}

	dto, err := h.service.ListMangas(ctx.Request.Context(), ur, &pq)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) GetMangaByID(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	mangaID, err := h.mangaIDFromPath(ctx)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	dto, err := h.service.GetMangaByID(ctx.Request.Context(), ur, mangaID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) UpdateManga(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	mangaID, err := h.mangaIDFromPath(ctx)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	var req service.UpdateMangaDTO
	if err := httptransport.BindJSON(ctx, &req, h.log); err != nil {
		h.handleError(ctx, err)
		return
	}

	dto, err := h.service.UpdateManga(ctx.Request.Context(), ur, mangaID, req)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) DeleteManga(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	mangaID, err := h.mangaIDFromPath(ctx)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	if err := h.service.DeleteManga(ctx.Request.Context(), ur, mangaID); err != nil {
		h.handleError(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, nil)
}
