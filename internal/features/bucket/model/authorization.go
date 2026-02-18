package model

import (
	"github.com/mairuu/mp-api/internal/app"
	a "github.com/mairuu/mp-api/internal/platform/authorization"
)

const (
	ResourceBucket a.Resource = "bucket"
)

const (
	ActionUpload a.Action = "upload"
)

func AllPolicies() []a.Policy {
	return a.Define(
		a.Grant(app.RoleAdmin).Regardless().On(ResourceBucket).Can(a.ActionAny),

		a.Grant(app.RoleUser).Regardless().On(ResourceBucket).Can(ActionUpload),
	)
}
