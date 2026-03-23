package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/features/bucket/model"
	"github.com/mairuu/mp-api/internal/platform/storage"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
)

func (h *BucketHandler) userRoleFromContext(ctx *gin.Context) *app.UserRole {
	return app.UserRoleFromContext(ctx)
}

func (h *BucketHandler) fail(ctx *gin.Context, err error) bool {
	if err != nil {
		h.handleError(ctx, err)
		return true
	}
	return false
}

func (h *BucketHandler) handleError(ctx *gin.Context, err error) {
	httptransport.HandleError(ctx, err, h.log, domainErrStatusMap)
}

var domainErrStatusMap = map[string]int{
	storage.ErrObjectNotFound.Error():   http.StatusNotFound,
	model.ErrFileRequired.Error():       http.StatusBadRequest,
	model.ErrRefIDCountMismatch.Error(): http.StatusBadRequest,
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
