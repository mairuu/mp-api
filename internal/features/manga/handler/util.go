package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/manga/model"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	perrors "github.com/mairuu/mp-api/internal/platform/errors"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
	"github.com/mairuu/mp-api/internal/platform/transport/http/middleware"
)

func (h *Handler) mangaIDFromPath(ctx *gin.Context) (uuid.UUID, error) {
	return uuidFromPath(ctx, "manga_id")
}

func (h *Handler) chapterIDFromPath(ctx *gin.Context) (uuid.UUID, error) {
	return uuidFromPath(ctx, "chapter_id")
}

func (h *Handler) userRoleFromContext(ctx *gin.Context) *app.UserRole {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return (&app.UserRole{}).OrGuest()
	}
	role, ok := middleware.GetUserRole(ctx)
	if !ok {
		return (&app.UserRole{}).OrGuest()
	}
	return (&app.UserRole{ID: userID, Role: authorization.Role(role)}).OrGuest()
}

func (h *Handler) handleError(ctx *gin.Context, err error) {
	code := toHTTPStatusCode(err)

	// for server errors, we log the error and return a generic error message to the client
	if code >= 500 {
		h.log.ErrorContext(ctx.Request.Context(), "internal server error", "error", err)
		httptransport.ErrorResponse(ctx, code, http.StatusText(code))
		return
	}

	// for client errors, we can return the error message to the client
	httptransport.ErrorResponse(ctx, code, err.Error())
}

func uuidFromPath(ctx *gin.Context, param string) (uuid.UUID, error) {
	id, ok := httptransport.GetParamAsUUID(ctx, param)
	if !ok {
		return uuid.Nil, httptransport.NewHandlerError(http.StatusBadRequest, "invalid "+param, nil)
	}
	return id, nil
}

var domainErrStatusMap = map[string]int{
	model.ErrCoverNotFound.Code:          http.StatusNotFound,
	model.ErrChapterNotFound.Code:        http.StatusNotFound,
	model.ErrMangaNotFound.Code:          http.StatusNotFound,
	model.ErrInvalidTitle.Code:           http.StatusBadRequest,
	model.ErrInvalidStatus.Code:          http.StatusBadRequest,
	model.ErrInvalidVolume.Code:          http.StatusBadRequest,
	model.ErrChapterAlreadyExists.Code:   http.StatusBadRequest,
	model.ErrVolumeAlreadyExists.Code:    http.StatusBadRequest,
	model.ErrPageNotFound.Code:           http.StatusBadRequest, // page not found can be caused by invalid staging object name or the staging object does not belong to the user
	model.ErrUnsupportedImageFormat.Code: http.StatusBadRequest,
}

func toHTTPStatusCode(err error) int {
	var statusCoder interface {
		Status() int
	}
	if errors.As(err, &statusCoder) {
		return statusCoder.Status()
	}

	var domainErr *perrors.DomainError
	if errors.As(err, &domainErr) {
		if code, ok := domainErrStatusMap[domainErr.Code]; ok {
			return code
		}
	}

	return http.StatusInternalServerError
}
