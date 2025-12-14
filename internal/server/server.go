package server

import (
	"context"
	"fmt"
	stdhttp "net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server is a simple HTTP server for metrics.
type Server struct {
	server *stdhttp.Server
}

// New creates a new metrics server.
func New(host string, port int) (*Server, error) {
	mux := stdhttp.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &stdhttp.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}

	return &Server{
		server: srv,
	}, nil
}

// Start starts the server in a goroutine.
func (s *Server) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
