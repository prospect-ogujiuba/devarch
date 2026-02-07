package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/priz/devarch-api/internal/api/handlers"
	mw "github.com/priz/devarch-api/internal/api/middleware"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/nginx"
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/internal/project"
	"github.com/priz/devarch-api/internal/scanner"
	"github.com/priz/devarch-api/internal/sync"
)

func NewRouter(db *sql.DB, containerClient *container.Client, podmanClient *podman.Client, syncManager *sync.Manager, projectScanner *scanner.Scanner, nginxGenerator *nginx.Generator, projectController *project.Controller) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link", "X-Total-Count", "X-Page", "X-Per-Page", "X-Total-Pages"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	serviceHandler := handlers.NewServiceHandler(db, containerClient, podmanClient)
	categoryHandler := handlers.NewCategoryHandler(db, containerClient, podmanClient)
	statusHandler := handlers.NewStatusHandler(db, podmanClient, syncManager)
	registryHandler := handlers.NewRegistryHandler(db)
	wsHandler := handlers.NewWebSocketHandler(syncManager)
	projectHandler := handlers.NewProjectHandler(db, projectScanner, projectController)
	runtimeHandler := handlers.NewRuntimeHandler(containerClient, podmanClient)
	nginxHandler := handlers.NewNginxHandler(nginxGenerator, containerClient)
	stackHandler := handlers.NewStackHandler(db, containerClient)
	instanceHandler := handlers.NewInstanceHandler(db, containerClient)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.APIKeyAuth)
		r.Use(mw.RateLimit(10, 50))
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
				r.Put("/dependencies", serviceHandler.UpdateDependencies)
				r.Put("/healthcheck", serviceHandler.UpdateHealthcheck)
				r.Put("/labels", serviceHandler.UpdateLabels)
				r.Put("/domains", serviceHandler.UpdateDomains)

				r.Get("/files", serviceHandler.ListConfigFiles)
				r.Get("/files/*", serviceHandler.GetConfigFile)
				r.Put("/files/*", serviceHandler.PutConfigFile)
				r.Delete("/files/*", serviceHandler.DeleteConfigFile)

				r.Get("/image", registryHandler.GetImage)
				r.Get("/tags", registryHandler.GetTags)
				r.Get("/vulnerabilities", registryHandler.GetVulnerabilities)
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
			r.Post("/scan", projectHandler.Scan)
			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", projectHandler.Get)
				r.Get("/services", projectHandler.Services)
				r.Get("/status", projectHandler.Status)
				r.Post("/start", projectHandler.Start)
				r.Post("/stop", projectHandler.Stop)
				r.Post("/restart", projectHandler.Restart)
			})
		})

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

		r.Route("/stacks", func(r chi.Router) {
			r.Get("/", stackHandler.List)
			r.Post("/", stackHandler.Create)

			r.Get("/trash", stackHandler.ListTrash)
			r.Post("/trash/{name}/restore", stackHandler.Restore)
			r.Delete("/trash/{name}", stackHandler.PermanentDelete)

			r.Route("/{name}", func(r chi.Router) {
				r.Get("/", stackHandler.Get)
				r.Put("/", stackHandler.Update)
				r.Delete("/", stackHandler.Delete)

				r.Post("/enable", stackHandler.Enable)
				r.Post("/disable", stackHandler.Disable)
				r.Post("/clone", stackHandler.Clone)
				r.Post("/rename", stackHandler.Rename)

				r.Get("/delete-preview", stackHandler.DeletePreview)
				r.Get("/network", stackHandler.NetworkStatus)
				r.Get("/compose", stackHandler.Compose)

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

						r.Put("/ports", instanceHandler.UpdatePorts)
						r.Put("/volumes", instanceHandler.UpdateVolumes)
						r.Put("/env-vars", instanceHandler.UpdateEnvVars)
						r.Put("/labels", instanceHandler.UpdateLabels)
						r.Put("/domains", instanceHandler.UpdateDomains)
						r.Put("/healthcheck", instanceHandler.UpdateHealthcheck)
						r.Put("/dependencies", instanceHandler.UpdateDependencies)

						r.Get("/files", instanceHandler.ListConfigFiles)
						r.Get("/files/*", instanceHandler.GetConfigFile)
						r.Put("/files/*", instanceHandler.PutConfigFile)
						r.Delete("/files/*", instanceHandler.DeleteConfigFile)

						r.Get("/effective-config", instanceHandler.EffectiveConfig)
					})
				})
			})
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return r
}
