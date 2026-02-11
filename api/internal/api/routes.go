package api

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/priz/devarch-api/internal/api/handlers"
	mw "github.com/priz/devarch-api/internal/api/middleware"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/crypto"
	"github.com/priz/devarch-api/internal/nginx"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/internal/project"
	"github.com/priz/devarch-api/internal/proxy"
	"github.com/priz/devarch-api/internal/scanner"
	"github.com/priz/devarch-api/internal/security"
	"github.com/priz/devarch-api/internal/sync"
	"github.com/priz/devarch-api/pkg/registry"
)

func NewRouter(db *sql.DB, containerClient *container.Client, podmanClient *podman.Client, syncManager *sync.Manager, projectScanner *scanner.Scanner, nginxGenerator *nginx.Generator, projectController *project.Controller, registryManager *registry.Manager, cipher *crypto.Cipher, secMode security.Mode) http.Handler {
	r := chi.NewRouter()

	// Read stack import size limit from env var (default 256MB)
	importMaxBytes := int64(256 << 20)
	if envVal := os.Getenv("STACK_IMPORT_MAX_BYTES"); envVal != "" {
		if parsed, err := strconv.ParseInt(envVal, 10, 64); err == nil {
			importMaxBytes = parsed
		}
	}

	// Read ALLOWED_ORIGINS from env var (default wildcard for dev-friendly behavior)
	allowedOrigins := []string{"*"}
	if envVal := os.Getenv("ALLOWED_ORIGINS"); envVal != "" {
		allowedOrigins = strings.Split(envVal, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}
	allowCredentials := len(allowedOrigins) > 0 && allowedOrigins[0] != "*"

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// MaxBodySize removed from global scope - applied per-route instead
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link", "X-Total-Count", "X-Page", "X-Per-Page", "X-Total-Pages"},
		AllowCredentials: allowCredentials,
		MaxAge:           300,
	}))

	serviceHandler := handlers.NewServiceHandler(db, containerClient, podmanClient, cipher)
	categoryHandler := handlers.NewCategoryHandler(db, containerClient, podmanClient)
	statusHandler := handlers.NewStatusHandler(db, podmanClient, syncManager)
	registryHandler := handlers.NewRegistryHandler(db, registryManager)
	wsHandler := handlers.NewWebSocketHandler(syncManager, allowedOrigins, secMode)
	projectHandler := handlers.NewProjectHandler(db, projectScanner, projectController, containerClient)
	runtimeHandler := handlers.NewRuntimeHandler(containerClient, podmanClient)
	nginxHandler := handlers.NewNginxHandler(nginxGenerator, containerClient)
	stackHandler := handlers.NewStackHandler(db, containerClient)
	instanceHandler := handlers.NewInstanceHandler(db, containerClient, cipher)
	networkHandler := handlers.NewNetworkHandler(db, containerClient)
	proxyGenerator := proxy.NewGenerator(db)
	proxyHandler := handlers.NewProxyHandler(proxyGenerator)
	authHandler := handlers.NewAuthHandler()

	r.Post("/api/v1/auth/validate", authHandler.Validate)

	// Stack import with large body limit - registered before main route group to avoid 10MB default
	r.With(mw.APIKeyAuth(secMode), mw.RateLimit(10, 50), mw.MaxBodySize(importMaxBytes)).Post("/api/v1/stacks/import", stackHandler.ImportStack)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.APIKeyAuth(secMode))
		r.Use(mw.RateLimit(10, 50))
		r.Use(mw.MaxBodySize(10 << 20)) // Default 10MB for most routes
		r.Route("/services", func(r chi.Router) {
			r.Get("/", serviceHandler.List)
			r.Post("/", serviceHandler.Create)
			r.Post("/bulk", serviceHandler.Bulk)

			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", serviceHandler.Get)
				r.Put("/", serviceHandler.Update)
				r.Delete("/", serviceHandler.Delete)

				r.Post("/start", serviceHandler.Start)
				r.Post("/stop", serviceHandler.Stop)
				r.Post("/restart", serviceHandler.Restart)
				r.Post("/rebuild", serviceHandler.Rebuild)

				r.Get("/status", serviceHandler.Status)
				r.Get("/logs", serviceHandler.Logs)
				r.Get("/metrics", serviceHandler.Metrics)
				r.Get("/compose", serviceHandler.Compose)

				r.Get("/versions", serviceHandler.Versions)
				r.Get("/versions/{version}", serviceHandler.GetVersion)
				r.Post("/validate", serviceHandler.Validate)
				r.Post("/export", serviceHandler.Export)
				r.Post("/materialize", serviceHandler.Materialize)

				r.Put("/ports", serviceHandler.UpdatePorts)
				r.Put("/volumes", serviceHandler.UpdateVolumes)
				r.Put("/env-vars", serviceHandler.UpdateEnvVars)
				r.Put("/env-files", serviceHandler.UpdateEnvFiles)
				r.Put("/networks", serviceHandler.UpdateNetworks)
				r.Put("/config-mounts", serviceHandler.UpdateConfigMounts)
				r.Put("/dependencies", serviceHandler.UpdateDependencies)
				r.Put("/healthcheck", serviceHandler.UpdateHealthcheck)
				r.Put("/labels", serviceHandler.UpdateLabels)
				r.Put("/domains", serviceHandler.UpdateDomains)

				r.Get("/files", serviceHandler.ListConfigFiles)
				r.Get("/files/*", serviceHandler.GetConfigFile)
				r.Put("/files/*", serviceHandler.PutConfigFile)
				r.Delete("/files/*", serviceHandler.DeleteConfigFile)

				r.Get("/exports", serviceHandler.ListExports)
				r.Put("/exports", serviceHandler.UpdateExports)
				r.Get("/imports", serviceHandler.ListImports)
				r.Put("/imports", serviceHandler.UpdateImports)

				r.Get("/image", registryHandler.GetImage)
				r.Get("/tags", registryHandler.GetTags)
				r.Get("/vulnerabilities", registryHandler.GetVulnerabilities)

				r.Post("/proxy-config", proxyHandler.GenerateForService)
			})
		})

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", categoryHandler.List)
			r.Get("/{name}", categoryHandler.Get)
			r.Get("/{name}/services", categoryHandler.Services)
			r.Post("/{name}/start", categoryHandler.Start)
			r.Post("/{name}/stop", categoryHandler.Stop)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectHandler.List)
			r.Post("/", projectHandler.Create)
			r.Post("/scan", projectHandler.Scan)
			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", projectHandler.Get)
				r.Put("/", projectHandler.Update)
				r.Delete("/", projectHandler.Delete)
				r.Get("/services", projectHandler.Services)
				r.Get("/status", projectHandler.Status)
				r.Post("/start", projectHandler.Start)
				r.Post("/stop", projectHandler.Stop)
				r.Post("/restart", projectHandler.Restart)
				r.Post("/services/{service}/start", projectHandler.StartService)
				r.Post("/services/{service}/stop", projectHandler.StopService)
				r.Post("/services/{service}/restart", projectHandler.RestartService)

				r.Post("/proxy-config", proxyHandler.GenerateForProject)
			})
		})

		r.Get("/proxy/types", proxyHandler.ListTypes)

		r.Route("/nginx", func(r chi.Router) {
			r.Post("/generate", nginxHandler.GenerateAll)
			r.Post("/generate/{name}", nginxHandler.GenerateOne)
			r.Post("/reload", nginxHandler.Reload)
		})

		r.Get("/status", statusHandler.Overview)
		r.Post("/sync", statusHandler.TriggerSync)
		r.Get("/sync/jobs", statusHandler.SyncJobs)

		r.Get("/ws/status", wsHandler.Handle)

		r.Get("/runtime/status", runtimeHandler.Status)
		r.Post("/runtime/switch", runtimeHandler.Switch)
		r.Get("/socket/status", runtimeHandler.SocketStatus)
		r.Post("/socket/start", runtimeHandler.SocketStart)

		r.Route("/networks", func(r chi.Router) {
			r.Get("/", networkHandler.List)
			r.Post("/", networkHandler.Create)
			r.Post("/bulk-remove", networkHandler.BulkRemove)
			r.Delete("/{name}", networkHandler.Remove)
		})

		r.Route("/registries", func(r chi.Router) {
			r.Get("/", registryHandler.ListRegistries)
			r.Route("/{registry}", func(r chi.Router) {
				r.Get("/search", registryHandler.SearchImages)
				r.Get("/images/*", registryHandler.ImageRoute)
			})
		})

		r.Route("/stacks", func(r chi.Router) {
			r.Get("/", stackHandler.List)
			r.Post("/", stackHandler.Create)

			r.Get("/trash", stackHandler.ListTrash)
			r.Post("/trash/{name}/restore", stackHandler.Restore)
			r.Delete("/trash/{name}", stackHandler.PermanentDelete)

			// /import moved outside route group (line 68) to use 256MB limit instead of 10MB default

			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", stackHandler.Get)
				r.Put("/", stackHandler.Update)
				r.Delete("/", stackHandler.Delete)

				r.Post("/enable", stackHandler.Enable)
				r.Post("/disable", stackHandler.Disable)
				r.Post("/clone", stackHandler.Clone)
				r.Post("/rename", stackHandler.Rename)

				r.Post("/stop", stackHandler.Stop)
				r.Post("/start", stackHandler.Start)
				r.Post("/restart", stackHandler.Restart)

				r.Get("/delete-preview", stackHandler.DeletePreview)
				r.Get("/network", stackHandler.NetworkStatus)
				r.Post("/network", stackHandler.CreateNetwork)
				r.Delete("/network", stackHandler.RemoveNetwork)
				r.Get("/compose", stackHandler.Compose)
				r.Get("/plan", stackHandler.Plan)
				r.Post("/apply", stackHandler.Apply)
				r.Get("/export", stackHandler.ExportStack)
				r.Post("/proxy-config", proxyHandler.GenerateForStack)

				r.Post("/lock", stackHandler.GenerateLock)
				r.Post("/lock/validate", stackHandler.ValidateLock)
				r.Post("/lock/refresh", stackHandler.RefreshLock)

				r.Route("/instances", func(r chi.Router) {
					r.Get("/", instanceHandler.List)
					r.Post("/", instanceHandler.Create)

					r.Route("/{instance}", func(r chi.Router) {
						r.Get("/", instanceHandler.Get)
						r.Put("/", instanceHandler.Update)
						r.Delete("/", instanceHandler.Delete)

						r.Post("/duplicate", instanceHandler.Duplicate)
						r.Put("/rename", instanceHandler.Rename)
						r.Get("/delete-preview", instanceHandler.DeletePreview)

						r.Post("/stop", instanceHandler.Stop)
						r.Post("/start", instanceHandler.Start)
						r.Post("/restart", instanceHandler.Restart)

						r.Put("/ports", instanceHandler.UpdatePorts)
						r.Put("/volumes", instanceHandler.UpdateVolumes)
						r.Put("/env-vars", instanceHandler.UpdateEnvVars)
						r.Put("/env-files", instanceHandler.UpdateEnvFiles)
						r.Put("/networks", instanceHandler.UpdateNetworks)
						r.Put("/config-mounts", instanceHandler.UpdateConfigMounts)
						r.Put("/labels", instanceHandler.UpdateLabels)
						r.Put("/domains", instanceHandler.UpdateDomains)
						r.Put("/healthcheck", instanceHandler.UpdateHealthcheck)
						r.Put("/dependencies", instanceHandler.UpdateDependencies)

						r.Get("/resources", instanceHandler.GetResourceLimits)
						r.Put("/resources", instanceHandler.UpdateResourceLimits)

						r.Get("/files", instanceHandler.ListConfigFiles)
						r.Get("/files/*", instanceHandler.GetConfigFile)
						r.Put("/files/*", instanceHandler.PutConfigFile)
						r.Delete("/files/*", instanceHandler.DeleteConfigFile)

						r.Get("/effective-config", instanceHandler.EffectiveConfig)
					})
				})

				r.Get("/wires", stackHandler.ListWires)
				r.Post("/wires/resolve", stackHandler.ResolveWires)
				r.Post("/wires/cleanup", stackHandler.CleanupOrphanedWires)
				r.Post("/wires", stackHandler.CreateWire)
				r.Delete("/wires/{wireId}", stackHandler.DeleteWire)
			})
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return r
}
