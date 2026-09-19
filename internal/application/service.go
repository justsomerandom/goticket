// Package application coordinates domain operations and their durable side effects.
package application

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"goticket/internal/organization"
	"goticket/internal/ticket"
	"goticket/internal/user"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	CreateOrganization(context.Context, organization.Organization) (organization.Organization, error)
	CreateUser(context.Context, user.User) (user.User, error)
	CreateTicket(context.Context, ticket.Ticket, ticket.AuditEvent, Job) (ticket.Ticket, error)
	UpdateTicket(context.Context, ticket.Ticket, ticket.AuditEvent) (ticket.Ticket, error)
	AddComment(context.Context, ticket.Comment, ticket.AuditEvent, Job) (ticket.Comment, error)
	GetTicket(context.Context, uuid.UUID) (ticket.Ticket, error)
	ListTickets(context.Context, ticket.ListFilter) ([]ticket.Ticket, error)
	ListComments(context.Context, uuid.UUID) ([]ticket.Comment, error)
	ListAudit(context.Context, uuid.UUID, int, int) ([]ticket.AuditEvent, error)
}
type Job struct {
	Type    string
	Payload map[string]any
}
type Service struct {
	store Store
	newID func() uuid.UUID
}

func New(store Store) *Service { return &Service{store: store, newID: uuid.New} }
func (s *Service) CreateOrganization(c context.Context, name string) (organization.Organization, error) {
	if !organization.ValidName(name) {
		return organization.Organization{}, ticket.ErrInvalid
	}
	return s.store.CreateOrganization(c, organization.Organization{ID: s.newID(), Name: name})
}
func (s *Service) CreateUser(c context.Context, orgID uuid.UUID, email, name string) (user.User, error) {
	if !user.Valid(email, name) || orgID == uuid.Nil {
		return user.User{}, ticket.ErrInvalid
	}
	return s.store.CreateUser(c, user.User{ID: s.newID(), OrganizationID: orgID, Email: email, Name: name})
}
func (s *Service) CreateTicket(c context.Context, orgID, actor uuid.UUID, subject, description string, priority ticket.Priority) (ticket.Ticket, error) {
	t := ticket.Ticket{ID: s.newID(), OrganizationID: orgID, CreatedBy: actor, Subject: subject, Description: description, Status: ticket.StatusOpen, Priority: priority}
	if orgID == uuid.Nil || actor == uuid.Nil || ticket.ValidateCreate(subject, t.Status, priority) != nil {
		return t, ticket.ErrInvalid
	}
	e := event(s.newID(), t.ID, actor, "created", map[string]any{"status": t.Status, "priority": t.Priority})
	return s.store.CreateTicket(c, t, e, Job{Type: "ticket.created", Payload: map[string]any{"ticket_id": t.ID.String()}})
}
func (s *Service) Assign(c context.Context, id, actor uuid.UUID, assignee *uuid.UUID) (ticket.Ticket, error) {
	t, e := s.store.GetTicket(c, id)
	if e != nil {
		return t, e
	}
	t.AssigneeID = assignee
	return s.store.UpdateTicket(c, t, event(s.newID(), id, actor, "assigned", map[string]any{"assignee_id": assigneeString(assignee)}))
}
func (s *Service) SetStatus(c context.Context, id, actor uuid.UUID, status ticket.Status) (ticket.Ticket, error) {
	if !ticket.ValidStatus(status) {
		return ticket.Ticket{}, ticket.ErrInvalid
	}
	t, e := s.store.GetTicket(c, id)
	if e != nil {
		return t, e
	}
	old := t.Status
	t.Status = status
	return s.store.UpdateTicket(c, t, event(s.newID(), id, actor, "status_changed", map[string]any{"from": old, "to": status}))
}
func (s *Service) SetPriority(c context.Context, id, actor uuid.UUID, p ticket.Priority) (ticket.Ticket, error) {
	if !ticket.ValidPriority(p) {
		return ticket.Ticket{}, ticket.ErrInvalid
	}
	t, e := s.store.GetTicket(c, id)
	if e != nil {
		return t, e
	}
	old := t.Priority
	t.Priority = p
	return s.store.UpdateTicket(c, t, event(s.newID(), id, actor, "priority_changed", map[string]any{"from": old, "to": p}))
}
func (s *Service) AddComment(c context.Context, id, actor uuid.UUID, body string, internal bool) (ticket.Comment, error) {
	if ticket.ValidateComment(body) != nil {
		return ticket.Comment{}, ticket.ErrInvalid
	}
	cm := ticket.Comment{ID: s.newID(), TicketID: id, AuthorID: actor, Body: body, Internal: internal}
	return s.store.AddComment(c, cm, event(s.newID(), id, actor, "comment_added", map[string]any{"internal": internal}), Job{Type: "ticket.comment_added", Payload: map[string]any{"ticket_id": id.String()}})
}
func (s *Service) GetTicket(c context.Context, id uuid.UUID) (ticket.Ticket, error) {
	return s.store.GetTicket(c, id)
}
func (s *Service) ListTickets(c context.Context, f ticket.ListFilter) ([]ticket.Ticket, error) {
	if f.OrganizationID == uuid.Nil {
		return nil, ticket.ErrInvalid
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		return nil, ticket.ErrInvalid
	}
	return s.store.ListTickets(c, f)
}
func (s *Service) Comments(c context.Context, id uuid.UUID) ([]ticket.Comment, error) {
	return s.store.ListComments(c, id)
}
func (s *Service) Audit(c context.Context, id uuid.UUID, limit, offset int) ([]ticket.AuditEvent, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.store.ListAudit(c, id, limit, offset)
}
func event(id, tid, actor uuid.UUID, typ string, data map[string]any) ticket.AuditEvent {
	return ticket.AuditEvent{ID: id, TicketID: tid, ActorID: &actor, Type: typ, Data: data}
}
func assigneeString(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
}
