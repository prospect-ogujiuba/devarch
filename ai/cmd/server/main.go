package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/priz/devarch-ai/internal/api"
	"github.com/priz/devarch-ai/internal/llm"
	"github.com/priz/devarch-ai/internal/ramalama"
)

func main() {
	slog.Info("starting devarch-ai")

	manager := ramalama.NewManager()
	client := llm.NewClient(manager)

	apiURL := os.Getenv("DEVARCH_API_URL")
	if apiURL == "" {
		apiURL = "http://devarch-api:8080"
	}
	apiKey := os.Getenv("DEVARCH_API_KEY")

	contextBuilder := llm.NewContextBuilder(apiURL, apiKey)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager.StartIdleWatcher(ctx)

	router := api.NewRouter(ctx, manager, client, contextBuilder)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
