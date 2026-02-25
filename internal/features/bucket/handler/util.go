package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/bucket/model"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	perrors "github.com/mairuu/mp-api/internal/platform/errors"
	"github.com/mairuu/mp-api/internal/platform/storage"
	"github.com/mairuu/mp-api/internal/platform/transport/http/middleware"
)

func userRoleFromContext(ctx *gin.Context) *app.UserRole {
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

var domainErrStatusMap = map[string]int{
	storage.ErrObjectNotFound.Error():   http.StatusNotFound,
	model.ErrFileRequired.Error():       http.StatusBadRequest,
	model.ErrRefIDCountMismatch.Error(): http.StatusBadRequest,
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

func setCachingHeaders(ctx *gin.Context, meta *storage.ObjectMetadata) (etag string, lastMod time.Time) {
	etag = `"` + meta.Hash + `"`
	lastMod = meta.LastModified.UTC().Truncate(time.Second)

	ctx.Header("ETag", etag)
	ctx.Header("Last-Modified", lastMod.Format(http.TimeFormat))
	// fixme: immutable content
	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")

	if meta.ContentType != "" {
		ctx.Header("Content-Type", meta.ContentType)
	}
	if meta.Size > 0 {
		ctx.Header("Content-Length", strconv.FormatInt(meta.Size, 10))
	}

	return
}

func isNotModified(ctx *gin.Context, etag string, lastMod time.Time) bool {
	if inm := ctx.GetHeader("If-None-Match"); inm != "" {
		if inm == etag {
			ctx.Status(http.StatusNotModified)
			return true
		}
	}

	if ims := ctx.GetHeader("If-Modified-Since"); ims != "" {
		if t, err := time.Parse(http.TimeFormat, ims); err == nil {
			if !lastMod.After(t) {
				ctx.Status(http.StatusNotModified)
				return true
			}
		}
	}

	return false
}
