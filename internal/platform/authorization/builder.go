package authorization

const (
	ScopeAny   Scope = "*"
	ScopeOther Scope = "other"
)

// Policy represents a single policy rule
type Policy struct {
	Subject string
	Object  string
	Action  string
}

// PolicyDefinition represents a policy definition that can be used to generate multiple policies
type PolicyDefinition struct {
	Role     Role
	Scope    Scope
	Resource Resource
	Actions  []Action
}

type PolicyBuilder struct {
	role Role
}

type ScopedPolicyBuilder struct {
	role  Role
	scope Scope
}

type ResourcedPolicyBuilder struct {
	role     Role
	scope    Scope
	resource Resource
}

func Define(policies ...PolicyDefinition) []Policy {
	var result []Policy
	for _, pd := range policies {
		for _, action := range pd.Actions {
			result = append(result, Policy{
				Subject: pd.Role.String(),
				Object:  scopedResource(pd.Resource, pd.Scope),
				Action:  action.String(),
			})
		}
	}

	return result
}

func Grant(role Role) *PolicyBuilder {
	return &PolicyBuilder{role: role}
}

func (b *PolicyBuilder) As(scope Scope) *ScopedPolicyBuilder {
	return &ScopedPolicyBuilder{role: b.role, scope: scope}
}

// Regardless is a shorthand for As(ScopeAny)
func (b *PolicyBuilder) Regardless() *ScopedPolicyBuilder {
	return &ScopedPolicyBuilder{role: b.role, scope: ScopeAny}
}

func (b *ScopedPolicyBuilder) On(resource Resource) *ResourcedPolicyBuilder {
	return &ResourcedPolicyBuilder{
		role:     b.role,
		scope:    b.scope,
		resource: resource,
	}
}

func (b *ResourcedPolicyBuilder) Can(actions ...Action) PolicyDefinition {
	return PolicyDefinition{
		Role:     b.role,
		Scope:    b.scope,
		Resource: b.resource,
		Actions:  actions,
	}
}
