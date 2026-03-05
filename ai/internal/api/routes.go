package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/priz/devarch-ai/internal/embedding"
	"github.com/priz/devarch-ai/internal/llm"
	"github.com/priz/devarch-ai/internal/ramalama"
)

func NewRouter(ctx context.Context, manager *ramalama.Manager, client *llm.Client, contextBuilder *llm.ContextBuilder, embedClient *embedding.Client, embedStore *embedding.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	h := NewHandler(ctx, manager, client, contextBuilder, embedClient, embedStore)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api/v1/ai", func(r chi.Router) {
		r.Get("/status", h.Status)
		r.Post("/generate", h.Generate)
		r.Post("/chat", h.Chat)
		r.Post("/diagnose", h.Diagnose)
		r.Post("/model/pull", h.PullModel)
		r.Get("/models", h.ListModels)
		r.Post("/stop", h.Stop)

		r.Post("/embed", h.Embed)
		r.Post("/embed/index", h.IndexDocument)
		r.Post("/embed/search", h.SearchDocuments)
	})

	return r
}
