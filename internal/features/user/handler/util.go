package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/features/user/model"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
)

func (h *UserHandler) fail(ctx *gin.Context, err error) bool {
	if err != nil {
		h.handleErrors(ctx, err)
		return true
	}
	return false
}

func (h *UserHandler) handleErrors(ctx *gin.Context, err error) {
	httptransport.HandleError(ctx, err, h.log, domainErrStatusMap)
}

var domainErrStatusMap = map[string]int{
	model.ErrUserNotFound.Code:         http.StatusNotFound,
	model.ErrUserAlreadyExists.Code:    http.StatusConflict,
	model.ErrInvalidCredentials.Code:   http.StatusUnauthorized,
	model.ErrInvalidEmail.Code:         http.StatusBadRequest,
	model.ErrInvalidUsername.Code:      http.StatusBadRequest,
	model.ErrInvalidPassword.Code:      http.StatusBadRequest,
	model.ErrRefreshTokenNotFound.Code: http.StatusUnauthorized,
	model.ErrRefreshTokenExpired.Code:  http.StatusUnauthorized,
	model.ErrRefreshTokenRevoked.Code:  http.StatusUnauthorized,
}
