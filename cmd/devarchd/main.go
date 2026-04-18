package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/api"
	"github.com/prospect-ogujiuba/devarch/internal/appsvc"
	"github.com/prospect-ogujiuba/devarch/internal/cache"
	"github.com/prospect-ogujiuba/devarch/internal/events"
)

type config struct {
	workspaceRoots []string
	catalogRoots   []string
	listen         string
	cachePath      string
}

type stringSliceFlag []string

func (f *stringSliceFlag) String() string {
	if f == nil {
		return ""
	}
	return strings.Join(*f, ",")
}

func (f *stringSliceFlag) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("flag value must not be empty")
	}
	*f = append(*f, trimmed)
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	return serve(ctx, cfg, nil)
}

func parseConfig(args []string) (config, error) {
	cfg := config{listen: "127.0.0.1:7777"}
	fs := flag.NewFlagSet("devarchd", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Var((*stringSliceFlag)(&cfg.workspaceRoots), "workspace-root", "Repeatable workspace root scanned recursively for devarch.workspace.yaml")
	fs.Var((*stringSliceFlag)(&cfg.catalogRoots), "catalog-root", "Repeatable daemon catalog root for /api/catalog/templates")
	fs.StringVar(&cfg.listen, "listen", cfg.listen, "TCP listen address")
	fs.StringVar(&cfg.cachePath, "cache", "", "Optional SQLite cache file path")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if len(cfg.workspaceRoots) == 0 {
		return config{}, fmt.Errorf("at least one --workspace-root is required")
	}
	return cfg, nil
}

func serve(ctx context.Context, cfg config, ready chan<- string) error {
	bus := events.NewBus()
	var store cache.Store
	if cfg.cachePath != "" {
		sqliteStore, err := cache.NewSQLite(cfg.cachePath)
		if err != nil {
			return err
		}
		store = sqliteStore
		defer store.Close()
	}

	service, err := appsvc.New(appsvc.Config{
		WorkspaceRoots: cfg.workspaceRoots,
		CatalogRoots:   cfg.catalogRoots,
		EventBus:       bus,
		Cache:          store,
	})
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", cfg.listen)
	if err != nil {
		return err
	}
	defer listener.Close()
	if ready != nil {
		ready <- listener.Addr().String()
	}

	httpServer := &http.Server{Handler: api.NewServer(service)}
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	err = httpServer.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		<-shutdownDone
		return nil
	}
	return err
}
