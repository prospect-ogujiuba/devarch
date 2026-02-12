package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// contextKey is an unexported type for context keys to avoid collisions
type contextKey string

// LoggerCtxKey is the context key for storing request-scoped loggers
const LoggerCtxKey contextKey = "logger"

// SlogMiddleware creates middleware that logs HTTP requests with structured logging
func SlogMiddleware(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract request ID from chi middleware
			requestID := chimw.GetReqID(r.Context())

			// Create request-scoped logger with request_id
			logger := base.With("request_id", requestID)

			// Store logger in context
			ctx := context.WithValue(r.Context(), LoggerCtxKey, logger)

			// Wrap response writer to capture status and bytes
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			// Serve the request
			next.ServeHTTP(ww, r.WithContext(ctx))

			// Log after handler completes (RoutePattern only available after routing)
			routePattern := chi.RouteContext(ctx).RoutePattern()
			op := r.Method + " " + routePattern

			// Build log attributes
			attrs := []any{
				"op", op,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"bytes", ww.BytesWritten(),
			}

			// Add stack/instance params if present
			if stack := chi.URLParam(r, "name"); stack != "" {
				attrs = append(attrs, "stack", stack)
			}
			if instance := chi.URLParam(r, "instance"); instance != "" {
				attrs = append(attrs, "instance", instance)
			}

			logger.Info("request completed", attrs...)
		})
	}
}

// LoggerFromContext extracts the request-scoped logger from context
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerCtxKey).(*slog.Logger); ok {
		return logger
	}
	// Fallback to default logger if middleware was bypassed (e.g., tests)
	return slog.Default()
}
