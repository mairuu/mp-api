package authorization

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/google/uuid"
)

type Enforcer struct {
	casbin *casbin.Enforcer
}

func NewEnforcer() (*Enforcer, error) {
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		// should never happen since modelText is a constant, but handle it anyway
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	enforcer.AddFunction("subjectMatch", subjectMatchFn)
	enforcer.AddFunction("scopeMatch", scopeMatchFn)
	enforcer.AddFunction("actionMatch", actionMatchFn)

	return &Enforcer{
		casbin: enforcer,
	}, nil
}

func (e *Enforcer) Enforce(
	userID uuid.UUID,
	role Role,
	resource Resource,
	action Action,
	target ScopeResolvable,
) error {
	var scope Scope
	if target != nil {
		scope = target.ScopeResolver()(userID)
	}

	ok, err := e.casbin.Enforce(
		role.String(),
		scopedResource(resource, scope),
		action.String(),
	)
	if err != nil {
		return err
	}
	if !ok {
		return NewForbiddenError(userID.String(), resource.String(), action.String(), "policy deny")
	}

	return nil
}

func (e *Enforcer) AddPolicies(providers ...[]Policy) error {
	for _, provider := range providers {
		for _, p := range provider {
			ok, err := e.casbin.AddPolicy(p.Subject, p.Object, p.Action)
			if err != nil {
				return err
			}
			if !ok {
				return nil // policy already exists, skip
			}
		}
	}

	return nil
}

func scopedResource(resource Resource, scope Scope) string {
	return resource.String() + ":" + scope.String()
}

const modelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = subjectMatch(r.sub, p.sub) && scopeMatch(p.obj, r.obj) && actionMatch(p.act, r.act)
`

func subjectMatchFn(args ...any) (any, error) {
	policySub, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: subMatch: expected string for args[0], got %T", ErrInvalidArguments, args[0])
	}
	requestSub, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: subMatch: expected string for args[1], got %T", ErrInvalidArguments, args[1])
	}

	return subjectMatch(policySub, requestSub), nil
}

func subjectMatch(policySub, requestSub string) bool {
	return policySub == requestSub || policySub == string(RoleAny)
}

// scopeMatchFn matches the policy resource string against the request resource string.
// it handles two cases:
//
//  1. exact match:        p.obj == r.obj
//     "manga:owner" vs "manga:owner" → true
//
//  2. wildcard scope:     policy scope is ScopeAny ("*")
//     "manga:*"     vs "manga:owner" → true
//     "manga:*"     vs "manga:other" → true
//     "chapter:*"   vs "manga:owner" → false  (different base resource)
//
// Registered as "scopeMatch" in the Casbin enforcer.
func scopeMatchFn(args ...any) (any, error) {
	policyObj, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: scopeMatch: expected string for args[0], got %T", ErrInvalidArguments, args[0])
	}
	requestObj, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: scopeMatch: expected string for args[1], got %T", ErrInvalidArguments, args[1])
	}

	return scopeMatch(policyObj, requestObj), nil
}

func scopeMatch(policyObj, requestObj string) bool {
	if policyObj == requestObj {
		return true
	}

	policyBase, policyScope, found := strings.Cut(policyObj, ":")
	if !found {
		return false
	}

	requestBase, _, found := strings.Cut(requestObj, ":")
	if !found {
		return false
	}

	return policyBase == requestBase && policyScope == string(ScopeAny)
}

// actionMatchFn matches the policy action against the request action.
// it handles two cases:
//
//  1. exact match:        p.act == r.act
//     "delete" vs "delete" → true
//
//  2. wildcard action:    policy action is ActionAny ("*")
//     "*"      vs "delete" → true
//     "*"      vs "read"   → true
//
// registered as "actionMatch" in the Casbin enforcer.
func actionMatchFn(args ...any) (any, error) {
	policyAct, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: actionMatch: expected string for args[0], got %T", ErrInvalidArguments, args[0])
	}
	requestAct, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: actionMatch: expected string for args[1], got %T", ErrInvalidArguments, args[1])
	}

	return actionMatch(policyAct, requestAct), nil
}

func actionMatch(policyAct, requestAct string) bool {
	return policyAct == requestAct || policyAct == string(ActionAny)
}
