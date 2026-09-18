package domain

import "context"

type TicketRepository interface {
	FindTicketByID(ctx context.Context, id TicketID) (*Ticket, error)
	SaveTicket(ctx context.Context, ticket *Ticket) error
	ListTicketsByOrganization(ctx context.Context, organizationID OrganizationID) ([]*Ticket, error)
}

type UserRepository interface {
	FindUserByID(ctx context.Context, id UserID) (*User, error)
	SaveUser(ctx context.Context, user *User) error
	ListUsersByOrganization(ctx context.Context, organizationID OrganizationID) ([]*User, error)
}

type OrganizationRepository interface {
	FindOrganizationByID(ctx context.Context, id OrganizationID) (*Organization, error)
	SaveOrganization(ctx context.Context, organization *Organization) error
}

type CommentRepository interface {
	FindCommentByID(ctx context.Context, id CommentID) (*Comment, error)
	SaveComment(ctx context.Context, comment *Comment) error
	ListCommentsByTicket(ctx context.Context, ticketID TicketID) ([]*Comment, error)
}
