package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
)

const (
	UserIDKey = "auth.user_id"
	RoleKey   = "auth.role"
)

type TokenValidator interface {
	ValidateToken(tokenString string) (uuid.UUID, string, error)
}

func Auth(validator TokenValidator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// if no authorization header, proceed without setting user ID
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			httptransport.ErrorResponse(ctx, http.StatusUnauthorized, "invalid authorization header format")
			ctx.Abort()
			return
		}

		token := parts[1]
		userID, role, err := validator.ValidateToken(token)
		if err != nil {
			httptransport.ErrorResponse(ctx, http.StatusUnauthorized, "invalid or expired token")
			ctx.Abort()
			return
		}
		// set into gin context
		ctx.Set(UserIDKey, userID)
		ctx.Set(RoleKey, role)
		ctx.Next()
	}
}

// this checks if the user is authenticated
// should not be used. services should handle authorization logic based on user ID and role from context.
func RequiredAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(UserIDKey)
		if !exists || userID == uuid.Nil {
			httptransport.ErrorResponse(c, http.StatusUnauthorized, "user not authenticated")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}

	id, ok := userID.(uuid.UUID)
	return id, ok
}

func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(RoleKey)
	if !exists {
		return "", false
	}

	r, ok := role.(string)
	return r, ok
}
