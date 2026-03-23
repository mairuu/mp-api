package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
)

func (h *Handler) mangaIDFromPath(ctx *gin.Context) (uuid.UUID, error) {
	return uuidFromPath(ctx, "manga_id")
}

func (h *Handler) chapterIDFromPath(ctx *gin.Context) (uuid.UUID, error) {
	return uuidFromPath(ctx, "chapter_id")
}

func uuidFromPath(ctx *gin.Context, param string) (uuid.UUID, error) {
	id, ok := httptransport.GetParamAsUUID(ctx, param)
	if !ok {
		return uuid.Nil, httptransport.NewHandlerError(http.StatusBadRequest, "invalid "+param, nil)
	}
	return id, nil
}

func (h *Handler) userRoleFromContext(ctx *gin.Context) *app.UserRole {
	return app.UserRoleFromContext(ctx)
}

func (h *Handler) fail(ctx *gin.Context, err error) bool {
	if err != nil {
		h.handleError(ctx, err)
		return true
	}
	return false
}

func (h *Handler) handleError(ctx *gin.Context, err error) {
	httptransport.HandleError(ctx, err, h.log, domainErrStatusMap)
}

var domainErrStatusMap = map[string]int{
	model.ErrMangaNotFound.Code:          http.StatusNotFound,
	model.ErrMangaAlreadyExists.Code:     http.StatusConflict,
	model.ErrInvalidTitle.Code:           http.StatusBadRequest,
	model.ErrInvalidStatus.Code:          http.StatusBadRequest,
	model.ErrInvalidVolume.Code:          http.StatusBadRequest,
	model.ErrChapterNotFound.Code:        http.StatusNotFound,
	model.ErrChapterAlreadyExists.Code:   http.StatusConflict,
	model.ErrInvalidChapterNumber.Code:   http.StatusBadRequest,
	model.ErrVolumeAlreadyExists.Code:    http.StatusConflict,
	model.ErrMultiplePrimaryCovers.Code:  http.StatusConflict,
	model.ErrCoverNotFound.Code:          http.StatusNotFound,
	model.ErrUnsupportedImageFormat.Code: http.StatusBadRequest,
	model.ErrEmptyPages.Code:             http.StatusBadRequest,
	model.ErrPageNotFound.Code:           http.StatusNotFound,
	model.ErrInvalidPageWidth.Code:       http.StatusBadRequest,
	model.ErrInvalidPageHeight.Code:      http.StatusBadRequest,
	model.ErrEmptyPageObjectName.Code:    http.StatusBadRequest,
}
