package jobs

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"goticket/internal/storage/postgres"
	"io"
	"log/slog"
	"testing"
)

type fakeStore struct {
	job               postgres.Job
	failed, completed bool
}

func (f *fakeStore) ClaimJob(context.Context) (postgres.Job, error)     { return f.job, nil }
func (f *fakeStore) CompleteJob(context.Context, uuid.UUID) error       { f.completed = true; return nil }
func (f *fakeStore) FailJob(context.Context, postgres.Job, error) error { f.failed = true; return nil }
func TestOneFailureSchedulesRetry(t *testing.T) {
	f := &fakeStore{job: postgres.Job{ID: uuid.New(), Attempts: 1, MaxAttempts: 5}}
	w := Worker{Store: f, Log: slog.New(slog.NewTextHandler(io.Discard, nil)), Handle: func(context.Context, postgres.Job) error { return errors.New("sink unavailable") }}
	if e := w.One(context.Background()); e != nil {
		t.Fatal(e)
	}
	if !f.failed || f.completed {
		t.Fatal("job was not failed for retry")
	}
}
