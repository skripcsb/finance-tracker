package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vadim/finance-tracker/internal/auth"
	"github.com/vadim/finance-tracker/internal/config"
	"github.com/vadim/finance-tracker/internal/handler"
	"github.com/vadim/finance-tracker/internal/service"
	"github.com/vadim/finance-tracker/internal/storage"
)

func main() {
	cfgPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		logger.Error("cant load config", "err", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		logger.Error("db connection failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("connected to database")

	tokenMgr := auth.NewTokenManager(cfg.JWT.Secret)
	userStore := storage.NewUserStorage(pool)
	catStore := storage.NewCategoryStorage(pool)
	txnStore := storage.NewTransactionStorage(pool)

	authSvc := service.NewAuthService(userStore, catStore, tokenMgr)
	txnSvc := service.NewTransactionService(txnStore, catStore)

	authH := handler.NewAuthHandler(authSvc, logger)
	txnH := handler.NewTransactionHandler(txnSvc, logger)
	router := handler.SetupRoutes(authH, txnH, tokenMgr, logger)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		logger.Info("starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	logger.Info("stopping...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
