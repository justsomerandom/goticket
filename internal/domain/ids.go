package domain

type TicketID string
type UserID string
type OrganizationID string
type CommentID string

func (id TicketID) IsZero() bool {
	return id == ""
}

func (id UserID) IsZero() bool {
	return id == ""
}

func (id OrganizationID) IsZero() bool {
	return id == ""
}

func (id CommentID) IsZero() bool {
	return id == ""
}
