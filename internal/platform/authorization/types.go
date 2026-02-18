package authorization

import "github.com/google/uuid"

type Role string

const (
	RoleAny Role = "*"
)

func (r Role) String() string {
	return string(r)
}

type Resource string

func (r Resource) String() string {
	return string(r)
}

type Action string

const (
	ActionAny Action = "*"
)

func (a Action) String() string {
	return string(a)
}

type Scope string

func (s Scope) String() string {
	return string(s)
}

type ScopeResolver func(userID uuid.UUID) Scope

type ScopeResolvable interface {
	ScopeResolver() ScopeResolver
}
