package jobs

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"goticket/internal/storage/postgres"
	"log/slog"
	"time"
)

type Handler func(context.Context, postgres.Job) error
type Store interface {
	ClaimJob(context.Context) (postgres.Job, error)
	CompleteJob(context.Context, uuid.UUID) error
	FailJob(context.Context, postgres.Job, error) error
}
type Worker struct {
	Store        Store
	Log          *slog.Logger
	Handle       Handler
	PollInterval time.Duration
}

func (w Worker) Run(c context.Context) error {
	if w.PollInterval == 0 {
		w.PollInterval = time.Second
	}
	tick := time.NewTicker(w.PollInterval)
	defer tick.Stop()
	for {
		if e := w.One(c); e != nil && !errors.Is(e, pgx.ErrNoRows) {
			w.Log.Error("job processing failed", "error", e)
		}
		select {
		case <-c.Done():
			return c.Err()
		case <-tick.C:
		}
	}
}
func (w Worker) One(c context.Context) error {
	j, e := w.Store.ClaimJob(c)
	if errors.Is(e, pgx.ErrNoRows) {
		return nil
	}
	if e != nil {
		return e
	}
	if e = w.Handle(c, j); e != nil {
		return w.Store.FailJob(c, j, e)
	}
	return w.Store.CompleteJob(c, j.ID)
}
