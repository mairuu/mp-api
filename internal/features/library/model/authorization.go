package model

import (
	a "github.com/mairuu/mp-api/internal/platform/authorization"
)

// todo: define policies for library features
func AllPolicies() []a.Policy {
	return a.Define()
}
