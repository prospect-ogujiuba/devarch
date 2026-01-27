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
	"github.com/priz/devarch-api/internal/podman"
	"github.com/priz/devarch-api/internal/sync"
)

func NewRouter(db *sql.DB, containerClient *container.Client, podmanClient *podman.Client, syncManager *sync.Manager) http.Handler {
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
	categoryHandler := handlers.NewCategoryHandler(db, containerClient)
	statusHandler := handlers.NewStatusHandler(db, podmanClient, syncManager)
	registryHandler := handlers.NewRegistryHandler(db)
	wsHandler := handlers.NewWebSocketHandler(syncManager)
	runtimeHandler := handlers.NewRuntimeHandler(containerClient, podmanClient)

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

		r.Get("/status", statusHandler.Overview)
		r.Post("/sync", statusHandler.TriggerSync)
		r.Get("/sync/jobs", statusHandler.SyncJobs)

		r.Get("/ws/status", wsHandler.Handle)

		r.Get("/runtime/status", runtimeHandler.Status)
		r.Post("/runtime/switch", runtimeHandler.Switch)
		r.Get("/socket/status", runtimeHandler.SocketStatus)
		r.Post("/socket/start", runtimeHandler.SocketStart)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return r
}
