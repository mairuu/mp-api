package app

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	"github.com/mairuu/mp-api/internal/platform/transport/http/middleware"
)

const (
	RoleGuest authorization.Role = "guest"
	RoleUser  authorization.Role = "user"
	RoleAdmin authorization.Role = "admin"
)

func ValidRoles() []authorization.Role {
	return []authorization.Role{RoleGuest, RoleUser, RoleAdmin}
}

func IsValidRole(role authorization.Role) bool {
	switch role {
	case RoleGuest, RoleUser, RoleAdmin:
		return true
	default:
		return false
	}
}

type UserRole struct {
	ID   uuid.UUID
	Role authorization.Role
}

// OrGuest returns a UserRole with RoleGuest if the original UserRole is nil or has an invalid role.
func (ur *UserRole) OrGuest() *UserRole {
	// if user is not authenticated or has an invalid role, we will treat them as a guest with no ID and role,
	// so we can still enforce permissions for unauthenticated users.
	if ur == nil || !IsValidRole(ur.Role) {
		return &UserRole{
			ID:   uuid.Nil,
			Role: RoleGuest,
		}
	}
	return ur
}

func UserRoleFromContext(ctx *gin.Context) *UserRole {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return (&UserRole{}).OrGuest()
	}
	role, ok := middleware.GetUserRole(ctx)
	if !ok {
		return (&UserRole{}).OrGuest()
	}
	return (&UserRole{ID: userID, Role: authorization.Role(role)}).OrGuest()
}
