package authorization

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidArguments = errors.New("invalid arguments")
)

// ForbiddenError represents a detailed forbidden error
type ForbiddenError struct {
	UserID   string
	Resource string
	Action   string
	Reason   string
}

func (e *ForbiddenError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("forbidden: user %s cannot %s %s - %s", e.UserID, e.Action, e.Resource, e.Reason)
	}
	return fmt.Sprintf("forbidden: user %s cannot %s %s", e.UserID, e.Action, e.Resource)
}

func (e *ForbiddenError) Status() int {
	return http.StatusForbidden
}

// NewForbiddenError creates a new ForbiddenError
func NewForbiddenError(userID, resource, action, reason string) *ForbiddenError {
	return &ForbiddenError{
		UserID:   userID,
		Resource: resource,
		Action:   action,
		Reason:   reason,
	}
}
