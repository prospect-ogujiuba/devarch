package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/priz/devarch-ai/internal/api"
	"github.com/priz/devarch-ai/internal/embedding"
	"github.com/priz/devarch-ai/internal/llm"
	"github.com/priz/devarch-ai/internal/ramalama"
)

func main() {
	slog.Info("starting devarch-ai")

	manager := ramalama.NewManager()
	client := llm.NewClient(manager)

	embedManager := ramalama.NewEmbedManager()
	embedClient := embedding.NewClient(embedManager)

	var embedStore *embedding.Store
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			slog.Error("failed to ping database", "error", err)
			os.Exit(1)
		}
		embedStore = embedding.NewStore(db)
		slog.Info("embedding store connected")
	}

	apiURL := os.Getenv("DEVARCH_API_URL")
	if apiURL == "" {
		apiURL = "http://devarch-api:8080"
	}
	apiKey := os.Getenv("DEVARCH_API_KEY")

	contextBuilder := llm.NewContextBuilder(apiURL, apiKey)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager.StartIdleWatcher(ctx)
	embedManager.StartIdleWatcher(ctx)

	router := api.NewRouter(ctx, manager, client, contextBuilder, embedClient, embedStore)

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
