package httptransport

import (
	"io"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(ctx *gin.Context, statusCode int, data any) {
	ctx.JSON(statusCode, gin.H{"success": true, "data": data})
}

func ErrorResponse(ctx *gin.Context, statusCode int, err any) {
	ctx.JSON(statusCode, gin.H{"success": false, "error": err})
}

func StreamCopy(ctx *gin.Context, reader io.Reader) {
	_, _ = io.Copy(ctx.Writer, reader)
}
