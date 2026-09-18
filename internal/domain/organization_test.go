package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewOrganization(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	organization, err := NewOrganization("org-1", "Acme", now)
	if err != nil {
		t.Fatalf("NewOrganization() error = %v", err)
	}

	if organization.ID != "org-1" {
		t.Fatalf("ID = %q, want org-1", organization.ID)
	}
	if !organization.CreatedAt.Equal(now) || !organization.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps = (%v, %v), want %v", organization.CreatedAt, organization.UpdatedAt, now)
	}
}

func TestOrganizationValidate(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		organization Organization
		want         error
	}{
		{
			name:         "valid",
			organization: Organization{ID: "org-1", Name: "Acme", CreatedAt: now, UpdatedAt: now},
		},
		{
			name:         "missing name",
			organization: Organization{ID: "org-1", Name: " ", CreatedAt: now, UpdatedAt: now},
			want:         ErrMissingRequiredField,
		},
		{
			name:         "missing timestamp",
			organization: Organization{ID: "org-1", Name: "Acme", CreatedAt: now},
			want:         ErrMissingRequiredField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.organization.Validate()
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
