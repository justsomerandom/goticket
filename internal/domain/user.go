package domain

import "time"

type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleAgent    UserRole = "agent"
	UserRoleAdmin    UserRole = "admin"
)

func (r UserRole) Valid() bool {
	switch r {
	case UserRoleCustomer, UserRoleAgent, UserRoleAdmin:
		return true
	default:
		return false
	}
}

type User struct {
	ID             UserID
	OrganizationID OrganizationID
	Name           string
	Email          string
	Role           UserRole
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewUser(id UserID, organizationID OrganizationID, name string, email string, role UserRole, at time.Time) (*User, error) {
	user := &User{
		ID:             id,
		OrganizationID: organizationID,
		Name:           name,
		Email:          email,
		Role:           role,
		CreatedAt:      at,
		UpdatedAt:      at,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

func (u User) Validate() error {
	if err := required("id", u.ID.IsZero()); err != nil {
		return err
	}
	if err := required("organization_id", u.OrganizationID.IsZero()); err != nil {
		return err
	}
	if err := required("email", blank(u.Email)); err != nil {
		return err
	}
	if err := required("created_at", u.CreatedAt.IsZero()); err != nil {
		return err
	}
	if err := required("updated_at", u.UpdatedAt.IsZero()); err != nil {
		return err
	}
	if !u.Role.Valid() {
		return FieldError{Field: "role", Err: ErrInvalidUserRole}
	}

	return nil
}
