package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"user-aggregation/internal/config"
	"user-aggregation/internal/lib/logger"
	"user-aggregation/internal/repo"
	"user-aggregation/internal/repo/postgres"
	"user-aggregation/internal/server"
	"user-aggregation/internal/server/handlers"
	_ "user-aggregation/docs"
)

// @title User Aggregation API
// @version 1.0
// @description This is app for user aggregation based on REST API
// @contact.url  http://github.com/h4tecancel
// @BasePath /
// @schemes http
func main() {
	cfg := config.MustLoad()

	log, cleanup := logger.Init(cfg.App.Env)
	defer cleanup()
	log.Info("App is starting!")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	db, err := postgres.New(ctx, cfg.Storage.DBURL, 1) // 1 - maxconns
	if err != nil {
		log.Error("smth with init postgres", "err", err)
		return
	}
	if c, ok := any(db).(interface{ Close() error }); ok {
		defer func() {
			if err := c.Close(); err != nil {
				log.Error("failed to close db", "err", err)
			}
		}()
	}

	var repoIface repo.Repo = db
	h := handlers.New(log, repoIface)
	s := server.New(h)

	if err := s.Start(ctx,
		cfg.HTTPServer.Address,
		cfg.HTTPServer.IdleTimeout,
		cfg.HTTPServer.Timeout,
		cfg.HTTPServer.Timeout); err != nil {
		log.Error("smth with server", "err", err)
		return
	}
}