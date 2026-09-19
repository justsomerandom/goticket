package application

import (
	"context"
	"github.com/google/uuid"
	"goticket/internal/organization"
	"goticket/internal/ticket"
	"goticket/internal/user"
	"testing"
)

type memoryStore struct {
	audit ticket.AuditEvent
	job   Job
}

func (m *memoryStore) CreateOrganization(_ context.Context, o organization.Organization) (organization.Organization, error) {
	return o, nil
}
func (m *memoryStore) CreateUser(_ context.Context, u user.User) (user.User, error) { return u, nil }
func (m *memoryStore) CreateTicket(_ context.Context, _ ticket.Ticket, a ticket.AuditEvent, j Job) (ticket.Ticket, error) {
	m.audit = a
	m.job = j
	return ticket.Ticket{}, nil
}
func (m *memoryStore) UpdateTicket(context.Context, ticket.Ticket, ticket.AuditEvent) (ticket.Ticket, error) {
	return ticket.Ticket{}, nil
}
func (m *memoryStore) AddComment(context.Context, ticket.Comment, ticket.AuditEvent, Job) (ticket.Comment, error) {
	return ticket.Comment{}, nil
}
func (m *memoryStore) GetTicket(context.Context, uuid.UUID) (ticket.Ticket, error) {
	return ticket.Ticket{}, nil
}
func (m *memoryStore) ListTickets(context.Context, ticket.ListFilter) ([]ticket.Ticket, error) {
	return nil, nil
}
func (m *memoryStore) ListComments(context.Context, uuid.UUID) ([]ticket.Comment, error) {
	return nil, nil
}
func (m *memoryStore) ListAudit(context.Context, uuid.UUID, int, int) ([]ticket.AuditEvent, error) {
	return nil, nil
}
func TestCreateTicketProducesAuditAndNotification(t *testing.T) {
	m := &memoryStore{}
	_, e := New(m).CreateTicket(context.Background(), uuid.New(), uuid.New(), "Cannot sign in", "", ticket.PriorityHigh)
	if e != nil {
		t.Fatal(e)
	}
	if m.audit.Type != "created" || m.job.Type != "ticket.created" {
		t.Fatalf("effects = %q, %q", m.audit.Type, m.job.Type)
	}
}
