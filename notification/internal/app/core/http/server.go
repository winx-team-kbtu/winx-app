package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"winx-notification/pkg/graylog/logger"
)

type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

const (
	_defaultReadTimeout       = 10 * time.Minute
	_defaultReadHeaderTimeout = 60 * time.Second
	_defaultWriteTimeout      = 10 * time.Minute
	_defaultAddr              = ":80"
	_defaultShutdownTimeout   = 3 * time.Second
)

func NewHttpServer(handler http.Handler, opts ...Option) *Server {
	httpServer := &http.Server{
		Handler:           handler,
		ReadTimeout:       _defaultReadTimeout,
		ReadHeaderTimeout: _defaultReadHeaderTimeout,
		WriteTimeout:      _defaultWriteTimeout,
	}

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.start()

	return s
}

func (s *Server) start() {
	go func() {
		fmt.Println("Starting HTTP server on", s.server.Addr)
		logger.Log.Println("Starting HTTP server on", s.server.Addr)
		s.notify <- s.server.ListenAndServe()

		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
