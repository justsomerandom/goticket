package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"goticket/internal/config"
	"goticket/internal/jobs"
	"goticket/internal/storage/postgres"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, e := config.Load()
	if e != nil {
		log.Error("configuration", "error", e)
		os.Exit(1)
	}
	pool, e := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if e != nil {
		log.Error("database", "error", e)
		os.Exit(1)
	}
	defer pool.Close()
	c, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	w := jobs.Worker{Store: postgres.New(pool), Log: log, Handle: func(_ context.Context, j postgres.Job) error {
		log.Info("notification delivered", "type", j.Type, "job_id", j.ID)
		return nil
	}}
	if e = w.Run(c); e != nil && e != context.Canceled {
		log.Error("worker stopped", "error", e)
	}
}
