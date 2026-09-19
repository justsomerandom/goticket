package ticket

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalid = errors.New("invalid ticket input")

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
	StatusClosed     Status = "closed"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type Ticket struct {
	ID, OrganizationID   uuid.UUID  `json:"id"`
	Number               int64      `json:"number"`
	Subject, Description string     `json:"subject"`
	Status               Status     `json:"status"`
	Priority             Priority   `json:"priority"`
	AssigneeID           *uuid.UUID `json:"assignee_id,omitempty"`
	CreatedBy            uuid.UUID  `json:"created_by"`
	CreatedAt, UpdatedAt time.Time  `json:"created_at"`
}
type Comment struct {
	ID, TicketID, AuthorID uuid.UUID `json:"id"`
	Body                   string    `json:"body"`
	Internal               bool      `json:"internal"`
	CreatedAt              time.Time `json:"created_at"`
}
type AuditEvent struct {
	ID, TicketID uuid.UUID      `json:"id"`
	ActorID      *uuid.UUID     `json:"actor_id,omitempty"`
	Type         string         `json:"type"`
	Data         map[string]any `json:"data"`
	CreatedAt    time.Time      `json:"created_at"`
}
type ListFilter struct {
	OrganizationID uuid.UUID
	Status         Status
	Priority       Priority
	AssigneeID     *uuid.UUID
	Query          string
	Limit, Offset  int
}

func ValidStatus(s Status) bool {
	return s == StatusOpen || s == StatusInProgress || s == StatusResolved || s == StatusClosed
}
func ValidPriority(p Priority) bool {
	return p == PriorityLow || p == PriorityNormal || p == PriorityHigh || p == PriorityUrgent
}
func ValidateCreate(subject string, status Status, priority Priority) error {
	if strings.TrimSpace(subject) == "" || len(subject) > 255 || !ValidStatus(status) || !ValidPriority(priority) {
		return ErrInvalid
	}
	return nil
}
func ValidateComment(body string) error {
	if strings.TrimSpace(body) == "" || len(body) > 10000 {
		return ErrInvalid
	}
	return nil
}
