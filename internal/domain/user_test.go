package domain

import (
	"errors"
	"testing"
	"time"
)

func TestUserRoleValid(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{name: "customer", role: UserRoleCustomer, want: true},
		{name: "agent", role: UserRoleAgent, want: true},
		{name: "admin", role: UserRoleAdmin, want: true},
		{name: "unknown", role: UserRole("owner"), want: false},
		{name: "zero", role: UserRole(""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	user, err := NewUser("user-1", "org-1", "Ada", "ada@example.com", UserRoleAgent, now)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	if user.ID != "user-1" || user.OrganizationID != "org-1" {
		t.Fatalf("user IDs = (%q, %q), want user-1/org-1", user.ID, user.OrganizationID)
	}
	if !user.CreatedAt.Equal(now) || !user.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps = (%v, %v), want %v", user.CreatedAt, user.UpdatedAt, now)
	}
}

func TestUserValidate(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		user User
		want error
	}{
		{
			name: "valid",
			user: User{ID: "user-1", OrganizationID: "org-1", Email: "ada@example.com", Role: UserRoleAdmin, CreatedAt: now, UpdatedAt: now},
		},
		{
			name: "missing email",
			user: User{ID: "user-1", OrganizationID: "org-1", Email: " ", Role: UserRoleAdmin, CreatedAt: now, UpdatedAt: now},
			want: ErrMissingRequiredField,
		},
		{
			name: "invalid role",
			user: User{ID: "user-1", OrganizationID: "org-1", Email: "ada@example.com", Role: UserRole("owner"), CreatedAt: now, UpdatedAt: now},
			want: ErrInvalidUserRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.want == nil {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tt.want)
			}
		})
	}
}
