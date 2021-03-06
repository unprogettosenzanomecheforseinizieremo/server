package main

import (
	"context"
	"github.com/metabs/server/internal/customer"
	customerHTTP "github.com/metabs/server/internal/customer/http"
	"github.com/metabs/server/internal/email"
	"github.com/metabs/server/internal/jwt"
	"github.com/metabs/server/internal/workspace"
	workspaceHTTP "github.com/metabs/server/internal/workspace/http"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/metabs/server/internal/log"

	database "github.com/metabs/server/internal/db"
	serverHTTP "github.com/metabs/server/internal/http"
	"github.com/metabs/server/internal/probe"
)

func main() {
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	logger, err := log.New()
	if err != nil {
		panic(err)
	}

	go func() {
		<-kill
		defer cancel()
		signal.Stop(kill)
		close(kill)
		logger.Info("Stopping the application...")
	}()

	logger.Info("Application stopped.")
	db, err := database.New(ctx)
	if err != nil {
		logger.With("error", err).Fatal("could not connect to the database.")
	}

	sv, err := jwt.New()
	if err != nil {
		logger.With("error", err).Fatal("could not create JWY signer verifier")
	}

	sender, err := email.New(logger)
	if err != nil {
		logger.With("error", err).Fatal("could not create email sender")
	}

	r := serverHTTP.NewRouter(logger)
	r.Route("/", probe.NewRouter(db, logger))
	r.Route("/workspaces", workspaceHTTP.NewRouter(&workspace.Repo{Client: db, Logger: logger}, sv, logger))
	r.Route("/customers", customerHTTP.NewRouter(&customer.Repo{Client: db, Logger: logger}, sv, sender, logger))
	srv := serverHTTP.New(r)

	done := make(chan struct{}, 1)
	go func(done chan<- struct{}) {
		<-ctx.Done()

		logger.Info("Stopping the serverHTTP...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.With("error", err).Fatal("Could not gracefully shutdown the serverHTTP.")
		}

		// If you have any metrics or logs that need to be read before the shut down, remove the comment to the next 3 lines
		// logger.Info("Waiting metrics and logger to be read.")
		// <-ctx.Done()
		// logger.Info("Metrics and logger should be read.")

		logger.Info("Server stopped.")
		close(done)
	}(done)

	logger.Info("Server running...")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.With("error", err).Fatal("Could not listen and serve.")
	}
	<-kill
}
