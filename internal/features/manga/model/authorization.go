package model

import (
	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	a "github.com/mairuu/mp-api/internal/platform/authorization"
)

const (
	ResourceManga a.Resource = "manga"
)

const (
	ActionCreate  a.Action = "create"
	ActionRead    a.Action = "read"
	ActionList    a.Action = "list"
	ActionUpdate  a.Action = "update"
	ActionDelete  a.Action = "delete"
	ActionPublish a.Action = "publish"
)

const (
	ScopeOwner a.Scope = "owner"
)

func AllPolicies() []a.Policy {
	return a.Define(
		a.Grant(app.RoleAdmin).Regardless().On(ResourceManga).Can(a.ActionAny),

		a.Grant(app.RoleGuest).Regardless().On(ResourceManga).Can(ActionRead, ActionList),

		a.Grant(app.RoleUser).Regardless().On(ResourceManga).Can(ActionCreate, ActionRead, ActionList),
		a.Grant(app.RoleUser).As(ScopeOwner).On(ResourceManga).Can(ActionUpdate, ActionDelete),
	)
}

func (m *Manga) ScopeResolver() a.ScopeResolver {
	return func(userID uuid.UUID) a.Scope {
		if m.OwnerID == userID {
			return ScopeOwner
		}
		return a.ScopeOther
	}
}
