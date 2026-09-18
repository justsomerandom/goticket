package domain

import "time"

type Organization struct {
	ID        OrganizationID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrganization(id OrganizationID, name string, at time.Time) (*Organization, error) {
	organization := &Organization{
		ID:        id,
		Name:      name,
		CreatedAt: at,
		UpdatedAt: at,
	}

	if err := organization.Validate(); err != nil {
		return nil, err
	}

	return organization, nil
}

func (o Organization) Validate() error {
	if err := required("id", o.ID.IsZero()); err != nil {
		return err
	}
	if err := required("name", blank(o.Name)); err != nil {
		return err
	}
	if err := required("created_at", o.CreatedAt.IsZero()); err != nil {
		return err
	}
	if err := required("updated_at", o.UpdatedAt.IsZero()); err != nil {
		return err
	}

	return nil
}
