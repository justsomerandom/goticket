// Package postgres contains the PostgreSQL implementation; SQL does not escape this package.
package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"goticket/internal/application"
	"goticket/internal/organization"
	"goticket/internal/ticket"
	"goticket/internal/user"
	"time"
)

const defaultJobLease = 5 * time.Minute

var ErrJobLeaseLost = errors.New("job processing lease lost")

type Store struct {
	pool     *pgxpool.Pool
	jobLease time.Duration
}

func New(pool *pgxpool.Pool) *Store { return NewWithJobLease(pool, defaultJobLease) }
func NewWithJobLease(pool *pgxpool.Pool, lease time.Duration) *Store {
	if lease <= 0 {
		lease = defaultJobLease
	}
	return &Store{pool: pool, jobLease: lease}
}
func (s *Store) Close()                       { s.pool.Close() }
func (s *Store) Ping(c context.Context) error { return s.pool.Ping(c) }
func (s *Store) CreateOrganization(c context.Context, o organization.Organization) (organization.Organization, error) {
	e := s.pool.QueryRow(c, `INSERT INTO organizations(id,name) VALUES($1,$2) RETURNING created_at`, o.ID, o.Name).Scan(&o.CreatedAt)
	return o, e
}
func (s *Store) CreateUser(c context.Context, u user.User) (user.User, error) {
	e := s.pool.QueryRow(c, `INSERT INTO users(id,organization_id,email,name) VALUES($1,$2,$3,$4) RETURNING created_at`, u.ID, u.OrganizationID, u.Email, u.Name).Scan(&u.CreatedAt)
	return u, translate(e)
}
func (s *Store) CreateTicket(c context.Context, t ticket.Ticket, a ticket.AuditEvent, j application.Job, key application.Idempotency) (ticket.Ticket, error) {
	e := s.withTx(c, func(tx pgx.Tx) error {
		created, e := insertTicket(c, tx, &t, key)
		if e != nil {
			return e
		}
		if !created {
			return nil
		}
		return insertEffects(c, tx, a, j)
	})
	return t, translate(e)
}
func (s *Store) UpdateTicket(c context.Context, t ticket.Ticket, a ticket.AuditEvent) (ticket.Ticket, error) {
	e := s.withTx(c, func(tx pgx.Tx) error {
		tag, e := tx.Exec(c, `UPDATE tickets SET status=$2,priority=$3,assignee_id=$4,updated_at=now() WHERE id=$1`, t.ID, t.Status, t.Priority, t.AssigneeID)
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 0 {
			return application.ErrNotFound
		}
		if e = tx.QueryRow(c, `SELECT updated_at FROM tickets WHERE id=$1`, t.ID).Scan(&t.UpdatedAt); e != nil {
			return e
		}
		return insertEffects(c, tx, a, application.Job{})
	})
	return t, translate(e)
}
func (s *Store) AddComment(c context.Context, cm ticket.Comment, a ticket.AuditEvent, j application.Job, key application.Idempotency) (ticket.Comment, error) {
	e := s.withTx(c, func(tx pgx.Tx) error {
		created, e := insertComment(c, tx, &cm, key)
		if e != nil {
			return e
		}
		if !created {
			return nil
		}
		return insertEffects(c, tx, a, j)
	})
	return cm, translate(e)
}

func insertTicket(c context.Context, tx pgx.Tx, t *ticket.Ticket, key application.Idempotency) (bool, error) {
	if key.Key == "" {
		e := tx.QueryRow(c, `INSERT INTO tickets(id,organization_id,subject,description,status,priority,created_by) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING number,created_at,updated_at`, t.ID, t.OrganizationID, t.Subject, t.Description, t.Status, t.Priority, t.CreatedBy).Scan(&t.Number, &t.CreatedAt, &t.UpdatedAt)
		return e == nil, e
	}
	e := tx.QueryRow(c, `INSERT INTO tickets(id,organization_id,subject,description,status,priority,created_by,idempotency_key,idempotency_hash) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT (organization_id,idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING RETURNING number,created_at,updated_at`, t.ID, t.OrganizationID, t.Subject, t.Description, t.Status, t.Priority, t.CreatedBy, key.Key, key.RequestHash).Scan(&t.Number, &t.CreatedAt, &t.UpdatedAt)
	if e == nil {
		return true, nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return false, e
	}
	var storedHash []byte
	e = tx.QueryRow(c, `SELECT id,organization_id,number,subject,description,status,priority,assignee_id,created_by,created_at,updated_at,idempotency_hash FROM tickets WHERE organization_id=$1 AND idempotency_key=$2`, t.OrganizationID, key.Key).Scan(&t.ID, &t.OrganizationID, &t.Number, &t.Subject, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &storedHash)
	if e != nil {
		return false, e
	}
	if !bytes.Equal(storedHash, key.RequestHash) {
		return false, application.ErrIdempotencyConflict
	}
	return false, nil
}

func insertComment(c context.Context, tx pgx.Tx, cm *ticket.Comment, key application.Idempotency) (bool, error) {
	if key.Key == "" {
		e := tx.QueryRow(c, `INSERT INTO comments(id,ticket_id,author_id,body,internal) VALUES($1,$2,$3,$4,$5) RETURNING created_at`, cm.ID, cm.TicketID, cm.AuthorID, cm.Body, cm.Internal).Scan(&cm.CreatedAt)
		return e == nil, e
	}
	e := tx.QueryRow(c, `INSERT INTO comments(id,ticket_id,author_id,body,internal,idempotency_key,idempotency_hash) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (ticket_id,idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING RETURNING created_at`, cm.ID, cm.TicketID, cm.AuthorID, cm.Body, cm.Internal, key.Key, key.RequestHash).Scan(&cm.CreatedAt)
	if e == nil {
		return true, nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return false, e
	}
	var storedHash []byte
	e = tx.QueryRow(c, `SELECT id,ticket_id,author_id,body,internal,created_at,idempotency_hash FROM comments WHERE ticket_id=$1 AND idempotency_key=$2`, cm.TicketID, key.Key).Scan(&cm.ID, &cm.TicketID, &cm.AuthorID, &cm.Body, &cm.Internal, &cm.CreatedAt, &storedHash)
	if e != nil {
		return false, e
	}
	if !bytes.Equal(storedHash, key.RequestHash) {
		return false, application.ErrIdempotencyConflict
	}
	return false, nil
}
func insertEffects(c context.Context, tx pgx.Tx, a ticket.AuditEvent, j application.Job) error {
	data, e := json.Marshal(a.Data)
	if e != nil {
		return e
	}
	if _, e = tx.Exec(c, `INSERT INTO audit_events(id,ticket_id,actor_id,type,data) VALUES($1,$2,$3,$4,$5)`, a.ID, a.TicketID, a.ActorID, a.Type, data); e != nil {
		return e
	}
	if j.Type != "" {
		p, e := json.Marshal(j.Payload)
		if e != nil {
			return e
		}
		_, e = tx.Exec(c, `INSERT INTO jobs(id,type,payload) VALUES($1,$2,$3)`, uuid.New(), j.Type, p)
		return e
	}
	return nil
}
func (s *Store) GetTicket(c context.Context, id uuid.UUID) (ticket.Ticket, error) {
	var t ticket.Ticket
	e := s.pool.QueryRow(c, `SELECT id,organization_id,number,subject,description,status,priority,assignee_id,created_by,created_at,updated_at FROM tickets WHERE id=$1`, id).Scan(&t.ID, &t.OrganizationID, &t.Number, &t.Subject, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	return t, translate(e)
}
func (s *Store) ListTickets(c context.Context, f ticket.ListFilter) ([]ticket.Ticket, error) {
	rows, e := s.pool.Query(c, `SELECT id,organization_id,number,subject,description,status,priority,assignee_id,created_by,created_at,updated_at FROM tickets WHERE organization_id=$1 AND ($2='' OR status=$2) AND ($3='' OR priority=$3) AND ($4::uuid IS NULL OR assignee_id=$4) AND ($5='' OR subject ILIKE '%' || $5 || '%' OR description ILIKE '%' || $5 || '%') ORDER BY updated_at DESC LIMIT $6 OFFSET $7`, f.OrganizationID, f.Status, f.Priority, f.AssigneeID, f.Query, f.Limit, f.Offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []ticket.Ticket{}
	for rows.Next() {
		var t ticket.Ticket
		if e = rows.Scan(&t.ID, &t.OrganizationID, &t.Number, &t.Subject, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Store) ListComments(c context.Context, id uuid.UUID) ([]ticket.Comment, error) {
	rows, e := s.pool.Query(c, `SELECT id,ticket_id,author_id,body,internal,created_at FROM comments WHERE ticket_id=$1 ORDER BY created_at`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []ticket.Comment{}
	for rows.Next() {
		var x ticket.Comment
		if e = rows.Scan(&x.ID, &x.TicketID, &x.AuthorID, &x.Body, &x.Internal, &x.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) ListAudit(c context.Context, id uuid.UUID, limit, offset int) ([]ticket.AuditEvent, error) {
	rows, e := s.pool.Query(c, `SELECT id,ticket_id,actor_id,type,data,created_at FROM audit_events WHERE ticket_id=$1 ORDER BY created_at LIMIT $2 OFFSET $3`, id, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []ticket.AuditEvent{}
	for rows.Next() {
		var x ticket.AuditEvent
		var d []byte
		if e = rows.Scan(&x.ID, &x.TicketID, &x.ActorID, &x.Type, &d, &x.CreatedAt); e != nil {
			return nil, e
		}
		if e = json.Unmarshal(d, &x.Data); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

type Job struct {
	ID                    uuid.UUID
	LockToken             uuid.UUID
	Type                  string
	Payload               map[string]any
	Attempts, MaxAttempts int
}

func (s *Store) ClaimJob(c context.Context) (Job, error) {
	var j Job
	var p []byte
	tx, e := s.pool.Begin(c)
	if e != nil {
		return j, e
	}
	defer tx.Rollback(c)
	lease := fmt.Sprintf("%f seconds", s.jobLease.Seconds())
	if _, e = tx.Exec(c, `UPDATE jobs SET state=CASE WHEN attempts>=max_attempts THEN 'dead' ELSE 'pending' END, locked_at=NULL, lock_token=NULL, next_attempt_at=now(), last_error='processing lease expired' WHERE state='running' AND locked_at < now()-$1::interval`, lease); e != nil {
		return j, e
	}
	j.LockToken = uuid.New()
	e = tx.QueryRow(c, `WITH next AS (SELECT id FROM jobs WHERE state='pending' AND next_attempt_at<=now() ORDER BY next_attempt_at FOR UPDATE SKIP LOCKED LIMIT 1) UPDATE jobs SET state='running',locked_at=now(),lock_token=$1,attempts=attempts+1 FROM next WHERE jobs.id=next.id RETURNING jobs.id,jobs.type,jobs.payload,jobs.attempts,jobs.max_attempts`, j.LockToken).Scan(&j.ID, &j.Type, &p, &j.Attempts, &j.MaxAttempts)
	if e != nil {
		return j, e
	}
	if e = json.Unmarshal(p, &j.Payload); e != nil {
		return j, e
	}
	return j, tx.Commit(c)
}
func (s *Store) CompleteJob(c context.Context, j Job) error {
	tag, e := s.pool.Exec(c, `UPDATE jobs SET state='completed',completed_at=now(),locked_at=NULL,lock_token=NULL WHERE id=$1 AND state='running' AND lock_token=$2`, j.ID, j.LockToken)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrJobLeaseLost
	}
	return nil
}
func (s *Store) FailJob(c context.Context, j Job, cause error) error {
	state := "pending"
	if j.Attempts >= j.MaxAttempts {
		state = "dead"
	}
	delay := time.Duration(1<<min(j.Attempts, 8)) * time.Second
	tag, e := s.pool.Exec(c, `UPDATE jobs SET state=$2,last_error=$3,locked_at=NULL,lock_token=NULL,next_attempt_at=now()+$4::interval WHERE id=$1 AND state='running' AND lock_token=$5`, j.ID, state, cause.Error(), fmt.Sprintf("%f seconds", delay.Seconds()), j.LockToken)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrJobLeaseLost
	}
	return nil
}
func (s *Store) withTx(c context.Context, fn func(pgx.Tx) error) error {
	tx, e := s.pool.Begin(c)
	if e != nil {
		return e
	}
	defer tx.Rollback(c)
	if e = fn(tx); e != nil {
		return e
	}
	return tx.Commit(c)
}
func translate(e error) error {
	if errors.Is(e, pgx.ErrNoRows) {
		return application.ErrNotFound
	}
	return e
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
