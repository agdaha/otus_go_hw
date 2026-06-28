package internalhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// Application will be expanded with business-logic methods in hw13.
type Application interface{}

type Server struct {
	server *http.Server
	logger Logger
	app    Application
}

func NewServer(logger Logger, app Application, host, port string) *Server {
	s := &Server{logger: logger, app: app}
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	s.server = &http.Server{
		Addr:              net.JoinHostPort(host, port),
		Handler:           loggingMiddleware(logger)(mux),
		ReadHeaderTimeout: 3 * time.Second,
	}
	return s
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hi!"))
}

func (s *Server) Start(_ context.Context) error {
	s.logger.Info("http server listening on " + s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
