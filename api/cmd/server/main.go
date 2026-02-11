package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/priz/devarch-api/internal/api"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/crypto"
	"github.com/priz/devarch-api/internal/nginx"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/internal/project"
	"github.com/priz/devarch-api/internal/scanner"
	"github.com/priz/devarch-api/internal/sync"
	"github.com/priz/devarch-api/pkg/registry"
	"github.com/priz/devarch-api/pkg/registry/dockerhub"
	"github.com/priz/devarch-api/pkg/registry/ghcr"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	key, err := crypto.LoadOrGenerateKey()
	if err != nil {
		log.Fatalf("failed to load encryption key: %v", err)
	}
	log.Println("encryption key loaded successfully")

	cipher := crypto.NewCipher(key)

	podmanClient, err := podman.NewClient()
	if err != nil {
		log.Fatalf("failed to create podman client: %v", err)
	}
	log.Printf("connected to podman socket: %s", podmanClient.SocketPath())

	if err := podmanClient.Ping(context.Background()); err != nil {
		log.Printf("warning: podman ping failed: %v", err)
	}

	containerClient, err := container.NewClient()
	if err != nil {
		log.Printf("warning: container client unavailable (compose operations disabled): %v", err)
	}

	projectController := project.NewController(db, containerClient)

	appsDir := os.Getenv("APPS_DIR")
	if appsDir == "" {
		appsDir = "/workspace/apps"
	}
	projectScanner := scanner.NewScanner(db, appsDir)

	if projects, err := projectScanner.ScanAndPersist(); err != nil {
		log.Printf("initial project scan failed: %v", err)
	} else {
		log.Println("initial project scan complete")
		for _, p := range projects {
			if p.ComposePath != "" {
				if err := projectController.EnsureStack(p.Name); err != nil {
					log.Printf("failed to import compose for project %s: %v", p.Name, err)
				}
			}
		}
		log.Println("project compose import complete")
	}

	nginxOutputDir := os.Getenv("NGINX_GENERATED_DIR")
	if nginxOutputDir == "" {
		nginxOutputDir = "/workspace/config/nginx/generated"
	}
	nginxGenerator := nginx.NewGenerator(db, nginxOutputDir)

	if err := nginxGenerator.GenerateAll(); err != nil {
		log.Printf("initial nginx generation failed: %v", err)
	} else {
		log.Println("initial nginx config generation complete")
	}

	registryManager := registry.NewManager()
	registryManager.Register(dockerhub.NewClient())
	registryManager.Register(ghcr.NewClient())

	syncManager := sync.NewManager(db, podmanClient, registryManager)
	router := api.NewRouter(db, containerClient, podmanClient, syncManager, projectScanner, nginxGenerator, projectController, registryManager, cipher)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	syncManager.Start(ctx)

	go func() {
		log.Printf("starting server on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}
