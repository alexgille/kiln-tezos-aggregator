package api

import (
	"context"
	"net/http"
	"time"
)

const GRACE_PERIOD = 10 * time.Second

type Server struct {
	http.Server
}

// NewServer returns a mono-endpoint REST API server listening to connections on the given
// address, and binding the given handler with the given route to it. Requests on any
// unsupported route will be responded to with HTTP-404.
func NewServer(addr, route string, hdl http.Handler) *Server {
	mux := http.NewServeMux()
	mux.Handle(route, hdl)
	return &Server{
		Server: http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

// Start starts the server until the context is cancelled.
// Ensures a grace period configured by GRACE_PERIOD to let
// pending requests been processed with interruption.
// Connections are dropped after that delay.
func (srv *Server) Start(ctx context.Context) error {
	cxlChan := make(chan error, 1)

	// run HTTP server in separate go routine
	go func() {
		cxlChan <- srv.ListenAndServe()
	}()

	// wait until server error or context cancellation
	select {
	case err := <-cxlChan:
		return err
	case <-ctx.Done():
		// graceful shutdown timeout
		tctx, cancel := context.WithTimeout(context.Background(), GRACE_PERIOD)
		defer cancel()
		return srv.Shutdown(tctx)
	}
}
