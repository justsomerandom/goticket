package domain

import "time"

type TicketStatus string

const (
	TicketStatusOpen     TicketStatus = "open"
	TicketStatusPending  TicketStatus = "pending"
	TicketStatusResolved TicketStatus = "resolved"
	TicketStatusClosed   TicketStatus = "closed"
)

func (s TicketStatus) Valid() bool {
	switch s {
	case TicketStatusOpen, TicketStatusPending, TicketStatusResolved, TicketStatusClosed:
		return true
	default:
		return false
	}
}

type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityNormal TicketPriority = "normal"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

func (p TicketPriority) Valid() bool {
	switch p {
	case TicketPriorityLow, TicketPriorityNormal, TicketPriorityHigh, TicketPriorityUrgent:
		return true
	default:
		return false
	}
}

type Ticket struct {
	ID             TicketID
	OrganizationID OrganizationID
	RequesterID    UserID
	AssigneeID     UserID
	Title          string
	Description    string
	Status         TicketStatus
	Priority       TicketPriority
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ResolvedAt     *time.Time
	ClosedAt       *time.Time
}

func NewTicket(id TicketID, organizationID OrganizationID, requesterID UserID, title string, description string, priority TicketPriority, at time.Time) (*Ticket, error) {
	ticket := &Ticket{
		ID:             id,
		OrganizationID: organizationID,
		RequesterID:    requesterID,
		Title:          title,
		Description:    description,
		Status:         TicketStatusOpen,
		Priority:       priority,
		CreatedAt:      at,
		UpdatedAt:      at,
	}

	if err := ticket.Validate(); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (t Ticket) Validate() error {
	if err := required("id", t.ID.IsZero()); err != nil {
		return err
	}
	if err := required("organization_id", t.OrganizationID.IsZero()); err != nil {
		return err
	}
	if err := required("requester_id", t.RequesterID.IsZero()); err != nil {
		return err
	}
	if err := required("title", blank(t.Title)); err != nil {
		return err
	}
	if err := required("created_at", t.CreatedAt.IsZero()); err != nil {
		return err
	}
	if err := required("updated_at", t.UpdatedAt.IsZero()); err != nil {
		return err
	}
	if !t.Status.Valid() {
		return FieldError{Field: "status", Err: ErrInvalidTicketStatus}
	}
	if !t.Priority.Valid() {
		return FieldError{Field: "priority", Err: ErrInvalidTicketPriority}
	}

	return nil
}

func (t Ticket) CanTransitionTo(next TicketStatus) bool {
	if !next.Valid() {
		return false
	}
	if t.Status == next {
		return true
	}

	switch t.Status {
	case TicketStatusOpen:
		return next == TicketStatusPending || next == TicketStatusResolved || next == TicketStatusClosed
	case TicketStatusPending:
		return next == TicketStatusOpen || next == TicketStatusResolved || next == TicketStatusClosed
	case TicketStatusResolved:
		return next == TicketStatusOpen || next == TicketStatusClosed
	case TicketStatusClosed:
		return false
	default:
		return false
	}
}

func (t *Ticket) TransitionTo(next TicketStatus, at time.Time) error {
	if t == nil {
		return FieldError{Field: "ticket", Err: ErrMissingRequiredField}
	}
	if !next.Valid() {
		return FieldError{Field: "status", Err: ErrInvalidTicketStatus}
	}
	if !t.CanTransitionTo(next) {
		return ErrInvalidStatusTransition
	}
	if err := required("updated_at", at.IsZero()); err != nil {
		return err
	}

	t.Status = next
	t.UpdatedAt = at

	switch next {
	case TicketStatusResolved:
		t.ResolvedAt = &at
	case TicketStatusClosed:
		t.ClosedAt = &at
		if t.ResolvedAt == nil {
			t.ResolvedAt = &at
		}
	case TicketStatusOpen, TicketStatusPending:
		t.ResolvedAt = nil
		t.ClosedAt = nil
	}

	return nil
}
