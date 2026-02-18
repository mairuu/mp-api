package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/features/user/model"
	"github.com/mairuu/mp-api/internal/features/user/service"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
	"github.com/mairuu/mp-api/internal/platform/transport/http/middleware"
)

type UserHandler struct {
	log     *slog.Logger
	service *service.Service
}

func NewUserHandler(logger *slog.Logger, service *service.Service) *UserHandler {
	return &UserHandler{log: logger, service: service}
}

func (h *UserHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)

	router.GET("/me", middleware.RequiredAuth(), h.GetMe)
}

func (h *UserHandler) Register(ctx *gin.Context) {
	var req service.RegisterDTO
	if err := httptransport.BindJSON(ctx, &req, h.log); err != nil {
		h.handleErrors(ctx, err)
		return
	}

	user, err := h.service.Register(ctx.Request.Context(), req)
	if err != nil {
		h.handleErrors(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusCreated, user)
}

func (h *UserHandler) Login(ctx *gin.Context) {
	var req service.LoginDTO
	if err := httptransport.BindJSON(ctx, &req, h.log); err != nil {
		h.handleErrors(ctx, err)
		return
	}

	response, err := h.service.Login(ctx.Request.Context(), req)
	if err != nil {
		h.handleErrors(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, response)
}

func (h *UserHandler) GetMe(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		httptransport.ErrorResponse(ctx, http.StatusUnauthorized, "user not authenticated")
		return
	}

	user, err := h.service.GetUserByID(ctx.Request.Context(), userID)
	if err != nil {
		h.handleErrors(ctx, err)
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, user)
}

func (h *UserHandler) handleErrors(ctx *gin.Context, err error) {
	code := h.toHTTPStatusCode(err)

	// server errors should not leak details to clients
	if code >= 500 {
		h.log.ErrorContext(ctx.Request.Context(), "internal server error", "error", err)
		httptransport.ErrorResponse(ctx, http.StatusInternalServerError, "internal server error")
		return
	}

	httptransport.ErrorResponse(ctx, code, err.Error())
}

func (h *UserHandler) toHTTPStatusCode(err error) int {
	var statusCoder interface {
		Status() int
	}
	if errors.As(err, &statusCoder) {
		return statusCoder.Status()
	}

	switch err {
	case model.ErrUserNotFound:
		return http.StatusNotFound
	case model.ErrUserAlreadyExists:
		return http.StatusConflict
	case model.ErrInvalidCredentials:
		return http.StatusUnauthorized
	case model.ErrInvalidEmail, model.ErrInvalidUsername, model.ErrInvalidPassword:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
