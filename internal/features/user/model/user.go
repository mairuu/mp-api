package model

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/mairuu/mp-api/internal/app"
	"github.com/mairuu/mp-api/internal/platform/authorization"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	Role         authorization.Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

func NewUser(username, email, passwordHash string) (*User, error) {
	now := time.Now()

	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePasswordHash(passwordHash); err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         app.RoleUser, // Default role
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// update builder
type UserUpdater struct {
	u    *User              // user to update
	opts []UserUpdateOption // options to apply
}

type UserUpdateOption func(*User) error

func (u *User) Updater() *UserUpdater {
	return &UserUpdater{u: u}
}

func (uu *UserUpdater) Email(email *string) *UserUpdater {
	uu.opts = append(uu.opts, func(u *User) error {
		if email == nil {
			return nil
		}
		if err := validateEmail(*email); err != nil {
			return err
		}
		u.Email = *email
		return nil
	})
	return uu
}

func validateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

func (uu *UserUpdater) Username(username *string) *UserUpdater {
	uu.opts = append(uu.opts, func(u *User) error {
		if username == nil {
			return nil
		}
		if err := validateUsername(*username); err != nil {
			return err
		}
		u.Username = *username
		return nil
	})
	return uu
}

func validateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

func (uu *UserUpdater) PasswordHash(passwordHash *string) *UserUpdater {
	uu.opts = append(uu.opts, func(u *User) error {
		if passwordHash == nil {
			return nil
		}
		if err := validatePasswordHash(*passwordHash); err != nil {
			return err
		}
		u.PasswordHash = *passwordHash
		return nil
	})
	return uu
}

func (uu *UserUpdater) Role(role *authorization.Role) *UserUpdater {
	uu.opts = append(uu.opts, func(u *User) error {
		if role == nil {
			return nil
		}
		if !app.IsValidRole(*role) {
			return ErrInvalidRole
		}
		u.Role = *role
		return nil
	})
	return uu
}

func validatePasswordHash(passwordHash string) error {
	if passwordHash == "" {
		return ErrInvalidPassword
	}
	return nil
}

// Apply applies the updates to the user.
// It validates each field and returns an error if any validation fails and rolls back the changes.
func (uu *UserUpdater) Apply() error {
	u := *uu.u
	for _, opt := range uu.opts {
		if err := opt(&u); err != nil {
			return err
		}
	}
	*uu.u = u
	uu.u.UpdatedAt = time.Now()
	return nil
}
