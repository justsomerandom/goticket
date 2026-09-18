package domain

import (
	"errors"
	"testing"
	"time"
)

func TestTicketStatusValid(t *testing.T) {
	tests := []struct {
		name   string
		status TicketStatus
		want   bool
	}{
		{name: "open", status: TicketStatusOpen, want: true},
		{name: "pending", status: TicketStatusPending, want: true},
		{name: "resolved", status: TicketStatusResolved, want: true},
		{name: "closed", status: TicketStatusClosed, want: true},
		{name: "unknown", status: TicketStatus("waiting"), want: false},
		{name: "zero", status: TicketStatus(""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTicketPriorityValid(t *testing.T) {
	tests := []struct {
		name     string
		priority TicketPriority
		want     bool
	}{
		{name: "low", priority: TicketPriorityLow, want: true},
		{name: "normal", priority: TicketPriorityNormal, want: true},
		{name: "high", priority: TicketPriorityHigh, want: true},
		{name: "urgent", priority: TicketPriorityUrgent, want: true},
		{name: "unknown", priority: TicketPriority("critical"), want: false},
		{name: "zero", priority: TicketPriority(""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTicket(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	ticket, err := NewTicket("ticket-1", "org-1", "user-1", "Printer is on fire", "Smoke everywhere.", TicketPriorityUrgent, now)
	if err != nil {
		t.Fatalf("NewTicket() error = %v", err)
	}

	if ticket.Status != TicketStatusOpen {
		t.Fatalf("Status = %q, want %q", ticket.Status, TicketStatusOpen)
	}
	if ticket.AssigneeID != "" {
		t.Fatalf("AssigneeID = %q, want zero value", ticket.AssigneeID)
	}
	if ticket.ResolvedAt != nil {
		t.Fatalf("ResolvedAt = %v, want nil", ticket.ResolvedAt)
	}
	if ticket.ClosedAt != nil {
		t.Fatalf("ClosedAt = %v, want nil", ticket.ClosedAt)
	}
	if !ticket.CreatedAt.Equal(now) || !ticket.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps = (%v, %v), want %v", ticket.CreatedAt, ticket.UpdatedAt, now)
	}
}

func TestNewTicketValidatesRequiredFields(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		make func() (*Ticket, error)
		want error
	}{
		{
			name: "missing id",
			make: func() (*Ticket, error) {
				return NewTicket("", "org-1", "user-1", "Title", "", TicketPriorityNormal, now)
			},
			want: ErrMissingRequiredField,
		},
		{
			name: "blank title",
			make: func() (*Ticket, error) {
				return NewTicket("ticket-1", "org-1", "user-1", "  ", "", TicketPriorityNormal, now)
			},
			want: ErrMissingRequiredField,
		},
		{
			name: "invalid priority",
			make: func() (*Ticket, error) {
				return NewTicket("ticket-1", "org-1", "user-1", "Title", "", TicketPriority("critical"), now)
			},
			want: ErrInvalidTicketPriority,
		},
		{
			name: "zero timestamp",
			make: func() (*Ticket, error) {
				return NewTicket("ticket-1", "org-1", "user-1", "Title", "", TicketPriorityNormal, time.Time{})
			},
			want: ErrMissingRequiredField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.make()
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tt.want)
			}
		})
	}
}

func TestTicketTransitionTo(t *testing.T) {
	createdAt := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	resolvedAt := createdAt.Add(time.Hour)

	ticket, err := NewTicket("ticket-1", "org-1", "user-1", "Title", "", TicketPriorityNormal, createdAt)
	if err != nil {
		t.Fatalf("NewTicket() error = %v", err)
	}

	if err := ticket.TransitionTo(TicketStatusResolved, resolvedAt); err != nil {
		t.Fatalf("TransitionTo(resolved) error = %v", err)
	}
	if ticket.Status != TicketStatusResolved {
		t.Fatalf("Status = %q, want %q", ticket.Status, TicketStatusResolved)
	}
	if ticket.ResolvedAt == nil || !ticket.ResolvedAt.Equal(resolvedAt) {
		t.Fatalf("ResolvedAt = %v, want %v", ticket.ResolvedAt, resolvedAt)
	}
	if ticket.ClosedAt != nil {
		t.Fatalf("ClosedAt = %v, want nil", ticket.ClosedAt)
	}
}

func TestTicketTransitionRules(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		from      TicketStatus
		to        TicketStatus
		wantError error
	}{
		{name: "open to pending", from: TicketStatusOpen, to: TicketStatusPending},
		{name: "pending to open", from: TicketStatusPending, to: TicketStatusOpen},
		{name: "resolved to closed", from: TicketStatusResolved, to: TicketStatusClosed},
		{name: "closed is terminal", from: TicketStatusClosed, to: TicketStatusOpen, wantError: ErrInvalidStatusTransition},
		{name: "invalid target", from: TicketStatusOpen, to: TicketStatus("waiting"), wantError: ErrInvalidTicketStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := &Ticket{
				ID:             "ticket-1",
				OrganizationID: "org-1",
				RequesterID:    "user-1",
				Title:          "Title",
				Status:         tt.from,
				Priority:       TicketPriorityNormal,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			err := ticket.TransitionTo(tt.to, now.Add(time.Hour))
			if tt.wantError == nil {
				if err != nil {
					t.Fatalf("TransitionTo() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tt.wantError)
			}
		})
	}
}

func TestTicketTransitionToClosedSetsResolutionIfNeeded(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	closedAt := now.Add(time.Hour)

	ticket, err := NewTicket("ticket-1", "org-1", "user-1", "Title", "", TicketPriorityNormal, now)
	if err != nil {
		t.Fatalf("NewTicket() error = %v", err)
	}

	if err := ticket.TransitionTo(TicketStatusClosed, closedAt); err != nil {
		t.Fatalf("TransitionTo(closed) error = %v", err)
	}
	if ticket.ClosedAt == nil || !ticket.ClosedAt.Equal(closedAt) {
		t.Fatalf("ClosedAt = %v, want %v", ticket.ClosedAt, closedAt)
	}
	if ticket.ResolvedAt == nil || !ticket.ResolvedAt.Equal(closedAt) {
		t.Fatalf("ResolvedAt = %v, want %v", ticket.ResolvedAt, closedAt)
	}
}
