package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// Declare an HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// shutdownError channel will receive any errors returned by the Shutdown() function
	shutdownError := make(chan error)

	// Start a background goroutine for listening signals.
	go func() {
		// quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1) // buffered channel with size 1

		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals
		// and relay them to the quit channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until a signal is received.
		s := <-quit

		app.logger.PrintInfo("shutting down server...", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	// Start the HTTP server
	app.logger.PrintInfo(fmt.Sprintf("starting %s server on %s", app.config.env, srv.Addr), map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Calling Shutdown() will cause http.ErrServerClosed error
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	// If graceful shutdown completely => log "stopped server" message
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
