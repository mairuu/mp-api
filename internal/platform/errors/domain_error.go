package errors

import (
	"fmt"
	"maps"
	"sort"
	"strings"
)

// DomainError represents a structured domain error with a code, optional message, and optional arguments.
type DomainError struct {
	Code    string
	Message string
	Args    map[string]string
}

// returns a structured string in the format:
// code=error_code; message=optional message; key1=value1; key2=value2
func (e *DomainError) Error() string {
	parts := []string{fmt.Sprintf("code=%s", e.Code)}

	if e.Message != "" {
		parts = append(parts, fmt.Sprintf("message=%s", e.Message))
	}

	if len(e.Args) > 0 {
		keys := make([]string, 0, len(e.Args))
		for k := range e.Args {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, e.Args[k]))
		}
	}

	return strings.Join(parts, "; ")
}

func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func New(code string) *DomainError {
	return &DomainError{Code: code}
}

// WithMessage adds a message to the DomainError and returns it for chaining.
func (e *DomainError) WithMessage(message string) *DomainError {
	e.Message = message
	return e
}

// WithArg adds a single argument to the DomainError and returns it for chaining.
func (e *DomainError) WithArg(key, value string) *DomainError {
	if e.Args == nil {
		e.Args = make(map[string]string)
	}
	e.Args[key] = value
	return e
}

// WithArgs adds multiple arguments to the DomainError and returns it for chaining.
func (e *DomainError) WithArgs(args map[string]string) *DomainError {
	if e.Args == nil {
		e.Args = make(map[string]string)
	}
	maps.Copy(e.Args, args)
	return e
}
