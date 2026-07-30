package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/api/router"
	"github.com/edsonmubezi/myapp/internal/config"
	"github.com/edsonmubezi/myapp/internal/container"
)

type Server struct {
	config    *config.Config
	container *container.Container
	db        interface{} // Store db reference for router initialization
}

func New(cfg *config.Config, container *container.Container) *Server {
	return &Server{
		config:    cfg,
		container: container,
		db:        container.DB, // Assuming container has DB field
	}
}

func (s *Server) Start() error {
	r := router.SetupRouter(s.container.UseCases, s.container.LogRepo, s.container.Storage, s.container.DB, s.container.AuditService, s.container.SecurityService, s.container.AlertingService, s.container.AppLogService)
	handler := middleware.CORSHandler(r)
	handler = middleware.Recoverer(handler)
	handler = middleware.RequestLogger(handler)
	addr := s.config.Address()

	// Optional TLS configuration
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")

	// Start with TLS if certificates are provided
	if certFile != "" && keyFile != "" {
		fmt.Printf("Starting HTTPS server on %s...\n", addr)
		fmt.Printf("TLS Certificate: %s\n", certFile)
		return http.ListenAndServeTLS(addr, certFile, keyFile, handler)
	}

	// Start HTTP server (TLS handled by reverse proxy in production)
	fmt.Printf("Starting HTTP server on %s...\n", addr)
	return http.ListenAndServe(addr, handler)
}

// StartWithGracefulShutdown starts the server with proper signal handling
// and graceful shutdown to prevent data loss during restarts/deployments
func (s *Server) StartWithGracefulShutdown() error {
	r := router.SetupRouter(s.container.UseCases, s.container.LogRepo, s.container.Storage, s.container.DB, s.container.AuditService, s.container.SecurityService, s.container.AlertingService, s.container.AppLogService)
	handler := middleware.CORSHandler(r)
	handler = middleware.Recoverer(handler)
	handler = middleware.RequestLogger(handler)
	addr := s.config.Address()

	// Configure HTTP server with timeouts
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // Increased for large payloads (PDF generation, file uploads)
		IdleTimeout:  120 * time.Second,
	}

	// Channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		certFile := os.Getenv("TLS_CERT_FILE")
		keyFile := os.Getenv("TLS_KEY_FILE")

		if certFile != "" && keyFile != "" {
			log.Printf("Starting HTTPS server on %s...\n", addr)
			serverErrors <- server.ListenAndServeTLS(certFile, keyFile)
		} else {
			log.Printf("Starting HTTP server on %s...\n", addr)
			serverErrors <- server.ListenAndServe()
		}
	}()

	// Channel to listen for OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		log.Printf("Received shutdown signal: %v\n", sig)
		log.Println("Starting graceful shutdown...")

		// Create context with timeout for graceful shutdown
		// Give outstanding requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v. Forcing shutdown...\n", err)
			if closeErr := server.Close(); closeErr != nil {
				return fmt.Errorf("forced shutdown error: %w", closeErr)
			}
			return fmt.Errorf("graceful shutdown timeout: %w", err)
		}

		log.Println("Graceful shutdown completed successfully")
	}

	return nil
}
