package model

import a "github.com/mairuu/mp-api/internal/platform/authorization"

func AllPolicies() []a.Policy {
	return a.Define()
}
