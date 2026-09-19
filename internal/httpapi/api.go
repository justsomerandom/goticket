package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"goticket/internal/application"
	"goticket/internal/ticket"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type API struct {
	service *application.Service
	ready   func(context.Context) error
	log     *slog.Logger
}

func New(s *application.Service, ready func(context.Context) error, log *slog.Logger) http.Handler {
	a := &API{s, ready, log}
	r := chi.NewRouter()
	r.Use(requestID, accessLog(log), recoverer)
	r.Get("/healthz", a.health)
	r.Get("/readyz", a.readyz)
	r.Route("/v1", func(r chi.Router) {
		r.Post("/organizations", a.createOrganization)
		r.Post("/organizations/{organizationID}/users", a.createUser)
		r.Post("/organizations/{organizationID}/tickets", a.createTicket)
		r.Get("/organizations/{organizationID}/tickets", a.listTickets)
		r.Get("/tickets/{ticketID}", a.getTicket)
		r.Patch("/tickets/{ticketID}/assignment", a.assign)
		r.Patch("/tickets/{ticketID}/status", a.status)
		r.Patch("/tickets/{ticketID}/priority", a.priority)
		r.Post("/tickets/{ticketID}/comments", a.comment)
		r.Get("/tickets/{ticketID}/comments", a.comments)
		r.Get("/tickets/{ticketID}/audit-events", a.audit)
	})
	return r
}
func (a *API) health(w http.ResponseWriter, r *http.Request) {
	write(w, 200, map[string]string{"status": "ok"})
}
func (a *API) readyz(w http.ResponseWriter, r *http.Request) {
	c, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if e := a.ready(c); e != nil {
		problem(w, 503, "not_ready", e.Error())
		return
	}
	write(w, 200, map[string]string{"status": "ready"})
}
func (a *API) createOrganization(w http.ResponseWriter, r *http.Request) {
	var x struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &x) {
		return
	}
	o, e := a.service.CreateOrganization(r.Context(), x.Name)
	respond(w, o, e, 201)
}
func (a *API) createUser(w http.ResponseWriter, r *http.Request) {
	org, ok := pathID(w, r, "organizationID")
	if !ok {
		return
	}
	var x struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if !decode(w, r, &x) {
		return
	}
	u, e := a.service.CreateUser(r.Context(), org, x.Email, x.Name)
	respond(w, u, e, 201)
}
func (a *API) createTicket(w http.ResponseWriter, r *http.Request) {
	org, ok := pathID(w, r, "organizationID")
	if !ok {
		return
	}
	var x struct {
		ActorID              string `json:"actor_id"`
		Subject, Description string
		Priority             ticket.Priority `json:"priority"`
	}
	if !decode(w, r, &x) {
		return
	}
	actor, e := uuid.Parse(x.ActorID)
	if e != nil {
		problem(w, 400, "validation_error", "actor_id must be a UUID")
		return
	}
	t, e := a.service.CreateTicket(r.Context(), org, actor, x.Subject, x.Description, x.Priority)
	respond(w, t, e, 201)
}
func (a *API) listTickets(w http.ResponseWriter, r *http.Request) {
	org, ok := pathID(w, r, "organizationID")
	if !ok {
		return
	}
	q := r.URL.Query()
	f := ticket.ListFilter{OrganizationID: org, Status: ticket.Status(q.Get("status")), Priority: ticket.Priority(q.Get("priority")), Query: q.Get("q"), Limit: queryInt(q.Get("limit"), 50), Offset: queryInt(q.Get("offset"), 0)}
	if x := q.Get("assignee_id"); x != "" {
		id, e := uuid.Parse(x)
		if e != nil {
			problem(w, 400, "validation_error", "assignee_id must be a UUID")
			return
		}
		f.AssigneeID = &id
	}
	v, e := a.service.ListTickets(r.Context(), f)
	respond(w, v, e, 200)
}
func (a *API) getTicket(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "ticketID")
	if ok {
		v, e := a.service.GetTicket(r.Context(), id)
		respond(w, v, e, 200)
	}
}
func (a *API) assign(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "ticketID")
	if !ok {
		return
	}
	var x struct {
		ActorID    string  `json:"actor_id"`
		AssigneeID *string `json:"assignee_id"`
	}
	if !decode(w, r, &x) {
		return
	}
	actor, e := uuid.Parse(x.ActorID)
	if e != nil {
		problem(w, 400, "validation_error", "actor_id must be a UUID")
		return
	}
	var assignee *uuid.UUID
	if x.AssigneeID != nil {
		v, e := uuid.Parse(*x.AssigneeID)
		if e != nil {
			problem(w, 400, "validation_error", "assignee_id must be a UUID")
			return
		}
		assignee = &v
	}
	v, e := a.service.Assign(r.Context(), id, actor, assignee)
	respond(w, v, e, 200)
}
func (a *API) status(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "ticketID")
	if !ok {
		return
	}
	var x struct {
		ActorID string        `json:"actor_id"`
		Status  ticket.Status `json:"status"`
	}
	if !decode(w, r, &x) {
		return
	}
	actor, e := uuid.Parse(x.ActorID)
	if e != nil {
		problem(w, 400, "validation_error", "actor_id must be a UUID")
		return
	}
	v, e := a.service.SetStatus(r.Context(), id, actor, x.Status)
	respond(w, v, e, 200)
}
func (a *API) priority(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "ticketID")
	if !ok {
		return
	}
	var x struct {
		ActorID  string          `json:"actor_id"`
		Priority ticket.Priority `json:"priority"`
	}
	if !decode(w, r, &x) {
		return
	}
	actor, e := uuid.Parse(x.ActorID)
	if e != nil {
		problem(w, 400, "validation_error", "actor_id must be a UUID")
		return
	}
	v, e := a.service.SetPriority(r.Context(), id, actor, x.Priority)
	respond(w, v, e, 200)
}
func (a *API) comment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "ticketID")
	if !ok {
		return
	}
	var x struct {
		ActorID  string `json:"actor_id"`
		Body     string `json:"body"`
		Internal bool   `json:"internal"`
	}
	if !decode(w, r, &x) {
		return
	}
	actor, e := uuid.Parse(x.ActorID)
	if e != nil {
		problem(w, 400, "validation_error", "actor_id must be a UUID")
		return
	}
	v, e := a.service.AddComment(r.Context(), id, actor, x.Body, x.Internal)
	respond(w, v, e, 201)
}
func (a *API) comments(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "ticketID")
	if ok {
		v, e := a.service.Comments(r.Context(), id)
		respond(w, v, e, 200)
	}
}
func (a *API) audit(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "ticketID")
	if ok {
		v, e := a.service.Audit(r.Context(), id, queryInt(r.URL.Query().Get("limit"), 50), queryInt(r.URL.Query().Get("offset"), 0))
		respond(w, v, e, 200)
	}
}
func pathID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	id, e := uuid.Parse(chi.URLParam(r, key))
	if e != nil {
		problem(w, 400, "validation_error", key+" must be a UUID")
		return uuid.Nil, false
	}
	return id, true
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		problem(w, 400, "invalid_json", e.Error())
		return false
	}
	return true
}
func respond(w http.ResponseWriter, v any, e error, status int) {
	if e == nil {
		write(w, status, v)
		return
	}
	if errors.Is(e, ticket.ErrInvalid) {
		problem(w, 400, "validation_error", e.Error())
	} else if errors.Is(e, application.ErrNotFound) {
		problem(w, 404, "not_found", "resource not found")
	} else {
		problem(w, 500, "internal_error", "internal server error")
	}
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func problem(w http.ResponseWriter, status int, code, msg string) {
	write(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg}, "request_id": requestIDFrom(w)})
}
func queryInt(x string, d int) int {
	if x == "" {
		return d
	}
	v, e := strconv.Atoi(x)
	if e != nil {
		return -1
	}
	return v
}

type key struct{}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), key{}, id)))
	})
}
func requestIDFrom(w http.ResponseWriter) string { return w.Header().Get("X-Request-ID") }
func accessLog(l *slog.Logger) func(http.Handler) http.Handler {
	return func(n http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			n.ServeHTTP(w, r)
			l.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
		})
	}
}
func recoverer(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				problem(w, 500, "internal_error", "internal server error")
			}
		}()
		n.ServeHTTP(w, r)
	})
}

var _ = strings.TrimSpace
