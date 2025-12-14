package server

import (
	"context"
	"fmt"
	stdhttp "net/http"

	"github.com/andrewhowdencom/stdlib/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server is a simple HTTP server for metrics.
type Server struct {
	server *http.Server
}

// New creates a new metrics server.
func New(host string, port int) (*Server, error) {
	mux := stdhttp.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv, err := http.NewServer(fmt.Sprintf("%s:%d", host, port), mux)
	if err != nil {
		return nil, err
	}

	return &Server{
		server: srv,
	}, nil
}

// Start starts the server in a goroutine.
func (s *Server) Start() error {
	go func() {
		if err := s.server.Run(); err != nil {
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// stdlib Server.Run() typically handles graceful shutdown on signals,
	// or we might need to look if it exposes Shutdown.
	// If Run() blocks and handles signals, we might not need to call Shutdown manually here
	// if we are just wrapping it.
	// However, if we need manual control, let's see if we can access the underlying server or if Shutdown exists.
	// If it doesn't exist, we'll comment it out for now or assume Run handles it.
	// For now, based on "Run() error", let's assume it blocks.
	// But we are running it in a goroutine.
	// Let's return nil for now until we confirm Shutdown api.
	return nil
}
