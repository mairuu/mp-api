package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
	"github.com/mairuu/mp-api/internal/platform/transport/http/middleware"
)

func (h *Handler) mangaIDFromPath(ctx *gin.Context) (uuid.UUID, error) {
	return uuidFromPath(ctx, "manga_id")
}

func uuidFromPath(ctx *gin.Context, param string) (uuid.UUID, error) {
	id, ok := httptransport.GetParamAsUUID(ctx, param)
	if !ok {
		return uuid.Nil, httptransport.NewHandlerError(http.StatusBadRequest, "invalid "+param, nil)
	}
	return id, nil
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

var domainErrStatusMap = map[string]int{}
