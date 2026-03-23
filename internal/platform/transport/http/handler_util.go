package httptransport

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	perrors "github.com/mairuu/mp-api/internal/platform/errors"
)

// ToHTTPStatusCode converts domain errors to HTTP status codes.
func ToHTTPStatusCode(err error, domainErrStatusMap map[string]int) int {
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

// HandleError responds with an appropriate HTTP status code and error message.
func HandleError(ctx *gin.Context, err error, log *slog.Logger, domainErrStatusMap map[string]int) {
	code := ToHTTPStatusCode(err, domainErrStatusMap)

	// for server errors, we log the error and return a generic error message to the client
	if code >= 500 {
		log.ErrorContext(ctx.Request.Context(), "internal server error", "error", err)
		ErrorResponse(ctx, code, http.StatusText(code))
		return
	}

	// for client errors, we can return the error message to the client
	ErrorResponse(ctx, code, err.Error())
}
