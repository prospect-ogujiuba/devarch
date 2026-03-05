package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/priz/devarch-api/internal/api"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/crypto"
	"github.com/priz/devarch-api/internal/llm"
	"github.com/priz/devarch-api/internal/nginx"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/internal/project"
	"github.com/priz/devarch-api/internal/ramalama"
	"github.com/priz/devarch-api/internal/scanner"
	"github.com/priz/devarch-api/internal/security"
	"github.com/priz/devarch-api/internal/sync"
	"github.com/priz/devarch-api/pkg/registry"
	"github.com/priz/devarch-api/pkg/registry/dockerhub"
	"github.com/priz/devarch-api/pkg/registry/ghcr"

	_ "github.com/lib/pq"
	_ "github.com/priz/devarch-api/docs"
)

// @title           DevArch API
// @version         1.0
// @description     Local microservices development environment API

// @host      localhost:8550
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key authentication header

func initLogger() *slog.Logger {
	var level slog.Level
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	programLevel := new(slog.LevelVar)
	programLevel.Set(level)

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: programLevel,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func main() {
	logger := initLogger()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable"
	}

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

	key, err := crypto.LoadOrGenerateKey()
	if err != nil {
		slog.Error("failed to load encryption key", "error", err)
		os.Exit(1)
	}
	slog.Info("encryption key loaded successfully")

	cipher := crypto.NewCipher(key)

	podmanClient, err := podman.NewClient()
	if err != nil {
		slog.Error("failed to create podman client", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to podman socket", "socket", podmanClient.SocketPath())

	if err := podmanClient.Ping(context.Background()); err != nil {
		slog.Warn("podman ping failed", "error", err)
	}

	containerClient, err := container.NewClient()
	if err != nil {
		slog.Warn("container client unavailable (compose operations disabled)", "error", err)
	}

	projectController := project.NewController(db, containerClient, cipher)

	appsDir := os.Getenv("APPS_DIR")
	if appsDir == "" {
		appsDir = "/workspace/apps"
	}
	projectScanner := scanner.NewScanner(db, appsDir)

	if projects, err := projectScanner.ScanAndPersist(); err != nil {
		slog.Error("initial project scan failed", "error", err)
	} else {
		slog.Info("initial project scan complete")
		for _, p := range projects {
			if p.ComposePath != "" {
				if err := projectController.EnsureStack(p.Name); err != nil {
					slog.Error("failed to import compose for project", "project", p.Name, "error", err)
				}
			}
		}
		slog.Info("project compose import complete")
	}

	nginxOutputDir := os.Getenv("NGINX_GENERATED_DIR")
	if nginxOutputDir == "" {
		nginxOutputDir = "/workspace/config/nginx/generated"
	}
	nginxGenerator := nginx.NewGenerator(db, nginxOutputDir)

	if err := nginxGenerator.GenerateAll(); err != nil {
		slog.Error("initial nginx generation failed", "error", err)
	} else {
		slog.Info("initial nginx config generation complete")
	}

	registryManager := registry.NewManager()
	registryManager.Register(dockerhub.NewClient())
	registryManager.Register(ghcr.NewClient())

	secMode, err := security.ParseMode(os.Getenv("SECURITY_MODE"))
	if err != nil {
		slog.Error("security configuration error", "error", err)
		os.Exit(1)
	}
	if err := security.ValidateConfig(secMode); err != nil {
		slog.Error("security configuration error", "error", err)
		os.Exit(1)
	}
	slog.Info("security mode", "mode", secMode)

	ramalamaManager := ramalama.NewManager(db)
	llmClient := llm.NewClient(ramalamaManager)

	syncManager := sync.NewManager(db, podmanClient, registryManager, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := api.NewRouter(ctx, db, containerClient, podmanClient, syncManager, projectScanner, nginxGenerator, projectController, registryManager, cipher, secMode, logger, ramalamaManager, llmClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:        ":" + port,
		Handler:     router,
		IdleTimeout: 60 * time.Second,
	}

	syncManager.Start(ctx)
	ramalamaManager.StartIdleWatcher(ctx)

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

	slog.Info("shutting down server")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
