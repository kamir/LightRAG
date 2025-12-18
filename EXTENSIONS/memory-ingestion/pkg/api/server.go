package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kamir/memory-connector/pkg/client"
	"github.com/kamir/memory-connector/pkg/config"
	"go.uber.org/zap"
)

// Server represents the HTTP API server
type Server struct {
	config         *config.Config
	logger         *zap.Logger
	httpServer     *http.Server
	lookupHandler  *LookupHandler
}

// NewServer creates a new API server
func NewServer(
	cfg *config.Config,
	memoryClient *client.MemoryClient,
	lightragClient *client.LightRAGClient,
	logger *zap.Logger,
) *Server {
	lookupHandler := NewLookupHandler(cfg, memoryClient, lightragClient, logger)

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"memory-connector","version":"0.1.0"}`)
	})

	// Lookup API endpoints
	mux.HandleFunc("/api/v1/lookup/by-entity", lookupHandler.HandleByEntity)
	mux.HandleFunc("/api/v1/lookup/by-memory", lookupHandler.HandleByMemory)
	mux.HandleFunc("/api/v1/lookup/resolve", lookupHandler.HandleResolveURI)
	mux.HandleFunc("/api/v1/lookup/parse-uris", lookupHandler.HandleParseURIs)

	// CORS and logging middleware
	handler := corsMiddleware(loggingMiddleware(mux, logger))

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		config:        cfg,
		logger:        logger,
		httpServer:    httpServer,
		lookupHandler: lookupHandler,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP API server",
		zap.String("address", s.httpServer.Addr),
	)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP API server")
	return s.httpServer.Shutdown(ctx)
}

// Middleware

func loggingMiddleware(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Int("status", rw.statusCode),
			zap.Duration("duration", duration),
		)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Response writer wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
