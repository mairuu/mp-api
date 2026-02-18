package app

import (
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/platform/authorization"
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
