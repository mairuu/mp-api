package httptransport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func GetParamAsUUID(ctx *gin.Context, param string) (uuid.UUID, bool) {
	idStr := ctx.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// obj must be a pointer to a struct
func BindJSON(ctx *gin.Context, obj any, log *slog.Logger) error {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		return recognizeBindError(err)
	}
	return nil
}

// obj must be a pointer to a struct
func BindQuery(ctx *gin.Context, obj any, log *slog.Logger) error {
	if err := ctx.ShouldBindQuery(obj); err != nil {
		return recognizeBindError(err)
	}
	return nil
}

// obj must be a pointer to a struct
func BindMultipartForm(ctx *gin.Context, obj any, log *slog.Logger) error {
	if err := ctx.ShouldBind(obj); err != nil {
		return recognizeBindError(err)
	}
	return nil
}

func recognizeBindError(err error) error {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) || errors.Is(err, io.EOF) {
		return NewHandlerError(http.StatusBadRequest, "", err)
	}
	var ute *json.UnmarshalTypeError
	if errors.As(err, &ute) {
		msg := fmt.Sprintf("expect value type %s for field %s; got %s", ute.Type, ute.Field, ute.Value)
		return NewHandlerError(http.StatusBadRequest, msg, nil)
	}
	return NewHandlerError(http.StatusInternalServerError, "internal server error", err)
}
