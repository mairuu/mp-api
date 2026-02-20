package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *DomainError
		want string
	}{
		{
			name: "code only",
			err:  &DomainError{Code: "manga_not_found"},
			want: "code=manga_not_found",
		},
		{
			name: "code with message",
			err: &DomainError{
				Code:    "manga_not_found",
				Message: "record not found in database",
			},
			want: "code=manga_not_found; message=record not found in database",
		},
		{
			name: "code with single arg",
			err: &DomainError{
				Code: "manga_not_found",
				Args: map[string]string{"id": "123"},
			},
			want: "code=manga_not_found; id=123",
		},
		{
			name: "code with message and args",
			err: &DomainError{
				Code:    "manga_not_found",
				Message: "record not found in database",
				Args:    map[string]string{"id": "123", "status": "active"},
			},
			want: "code=manga_not_found; message=record not found in database; id=123; status=active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.want, got, "DomainError.Error() did not produce expected string")
		})
	}
}

func TestDomainError_Is(t *testing.T) {
	err1 := &DomainError{Code: "manga_not_found"}
	err2 := &DomainError{Code: "manga_not_found", Message: "different message"}
	err3 := &DomainError{Code: "different_code"}
	stdErr := errors.New("standard error")

	assert.True(t, err1.Is(err2), "expected errors with same code to be equal")
	assert.False(t, err1.Is(err3), "expected errors with different codes to not be equal")
	assert.False(t, err1.Is(stdErr), "expected DomainError to not equal standard error")
}

func TestDomainError_Chaining(t *testing.T) {
	err := New("manga_not_found").
		WithMessage("record not found in database").
		WithArg("id", "123").
		WithArg("status", "active")

	want := "code=manga_not_found; message=record not found in database; id=123; status=active"

	assert.Equal(t, want, err.Error(), "chained DomainError did not produce expected error string")
}

func TestDomainError_WithArgs(t *testing.T) {
	err := New("manga_not_found").
		WithMessage("record not found").
		WithArgs(map[string]string{
			"id":     "123",
			"status": "active",
		})

	want := "code=manga_not_found; message=record not found; id=123; status=active"

	assert.Equal(t, want, err.Error(), "DomainError with WithArgs did not produce expected error string")
}
