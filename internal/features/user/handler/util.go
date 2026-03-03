package handler

import (
	"errors"
	"net/http"

	"github.com/mairuu/mp-api/internal/features/user/model"
	perrors "github.com/mairuu/mp-api/internal/platform/errors"
)

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

func (h *UserHandler) toHTTPStatusCode(err error) int {
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
