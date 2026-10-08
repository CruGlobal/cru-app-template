// Command server is a minimal Cru app: a plain net/http server with a health
// check, so the container builds and deploys as-is. Grow it into whatever you
// need (routes, a database, background work).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// shutdownBudget is how long in-flight requests get to finish after a stop
// signal. Cloud Run kills the container 10 seconds after SIGTERM and ECS after
// 30, so this fits inside both.
const shutdownBudget = 8 * time.Second

func main() {
	// JSON lines on stdout, which Datadog parses into fields.
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// run serves until SIGINT or SIGTERM, then drains in-flight requests.
func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen on $PORT (defaults to 8080). The platform routes traffic to that
	// port: an ALB target group on ECS, Cloud Run's own router on Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Info("listening", "port", port)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}
	// Back to the default signal handling, so a second Ctrl-C stops the process
	// at once instead of waiting out the drain.
	stop()

	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownBudget)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		// A request outlasted the budget. Cut it off: the platform is about to
		// kill the container anyway, and this is still a normal stop.
		log.Warn("requests still running after the shutdown budget; closing them")
		if err := srv.Close(); err != nil {
			return err
		}
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func routes() http.Handler {
	mux := http.NewServeMux()

	// Health check: the platform calls this to know the app is alive. Keep a 200
	// here working, or deploys are marked unhealthy.
	health := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /up", health)

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if email := signedInEmail(r); email != "" {
			_, _ = fmt.Fprintf(w, "Hello, %s 👋\n", email)
			return
		}
		_, _ = w.Write([]byte("Hello from your Cru app 👋\n"))
	})

	return requireIAP(mux)
}
