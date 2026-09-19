package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"goticket/internal/application"
	"goticket/internal/config"
	"goticket/internal/httpapi"
	"goticket/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, e := config.Load()
	if e != nil {
		log.Error("configuration", "error", e)
		os.Exit(1)
	}
	ctx := context.Background()
	pool, e := pgxpool.New(ctx, cfg.DatabaseURL)
	if e != nil {
		log.Error("database", "error", e)
		os.Exit(1)
	}
	store := postgres.New(pool)
	defer store.Close()
	srv := &http.Server{Addr: cfg.HTTPAddress, Handler: httpapi.New(application.New(store), store.Ping, log), ReadHeaderTimeout: 5 * time.Second}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Info("api listening", "address", cfg.HTTPAddress)
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			log.Error("server", "error", e)
		}
	}()
	<-stop
	c, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	srv.Shutdown(c)
}
