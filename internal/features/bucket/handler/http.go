package handler

import (
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/features/bucket/service"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
)

type BucketHandler struct {
	log     *slog.Logger
	service *service.Service
}

func NewBucketHandler(log *slog.Logger, service *service.Service) *BucketHandler {
	return &BucketHandler{log: log, service: service}
}

func (h *BucketHandler) RegisterRoutes(router gin.IRouter) {
	uploads := router.Group("/storage")
	{
		uploads.POST("/", h.Upload)

		// todo: make this configurable
		enableServing := false
		if enableServing {
			uploads.GET("/*object_name", h.Get)
			uploads.HEAD("/*object_name", h.Head)
		}
	}
}

// upload temporary files to the bucket, these files will be deleted after some period of time
func (h *BucketHandler) Upload(ctx *gin.Context) {
	ur := userRoleFromContext(ctx)

	form, err := ctx.MultipartForm()
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		h.handleError(ctx, httptransport.NewHandlerError(http.StatusBadRequest, "file is required", nil))
		return
	}

	result, err := h.service.UploadFiles(ctx.Request.Context(), ur, files)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusCreated, result)
}

func (h *BucketHandler) Get(ctx *gin.Context) {
	objectName := strings.TrimPrefix(ctx.Param("object_name"), "/")
	if !shouldServeObject(objectName) {
		ctx.Status(http.StatusNotFound)
		return
	}

	meta, err := h.service.GetMetadata(ctx.Request.Context(), objectName)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	etag, lastMod := setCachingHeaders(ctx, meta)

	if isNotModified(ctx, etag, lastMod) {
		return
	}

	reader, err := h.service.Download(ctx.Request.Context(), objectName)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	defer reader.Close()

	http.ServeContent(
		ctx.Writer,
		ctx.Request,
		path.Base(objectName),
		lastMod,
		reader,
	)
}

func (h *BucketHandler) Head(ctx *gin.Context) {
	objectName := strings.TrimPrefix(ctx.Param("object_name"), "/")
	if !shouldServeObject(objectName) {
		ctx.Status(http.StatusNotFound)
		return
	}

	meta, err := h.service.GetMetadata(ctx.Request.Context(), objectName)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	etag, lastMod := setCachingHeaders(ctx, meta)

	if isNotModified(ctx, etag, lastMod) {
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *BucketHandler) handleError(ctx *gin.Context, err error) {
	code := toHTTPStatusCode(err)

	if code >= 500 {
		h.log.Error("bucket handler error", "error", err)
	}

	ctx.Status(code)
}
