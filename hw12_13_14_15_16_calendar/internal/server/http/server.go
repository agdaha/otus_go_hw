package internalhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
)

type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// Application will be expanded with business-logic methods in hw13.
type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListDayEvents(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListWeekEvents(ctx context.Context, start time.Time) ([]storage.Event, error)
	ListMonthEvents(ctx context.Context, start time.Time) ([]storage.Event, error)
}

type Server struct {
	server *http.Server
	logger Logger
	app    Application
}

func NewServer(logger Logger, app Application, host, port string) *Server {
	s := &Server{logger: logger, app: app}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /events", s.handleCreateEvent)
	mux.HandleFunc("PUT /events", s.handleUpdateEvent)
	mux.HandleFunc("DELETE /events/{id}", s.handleDeleteEvent)
	mux.HandleFunc("GET /events/day", s.handleListDayEvents)
	mux.HandleFunc("GET /events/week", s.handleListWeekEvents)
	mux.HandleFunc("GET /events/month", s.handleListMonthEvents)

	s.server = &http.Server{
		Addr:              net.JoinHostPort(host, port),
		Handler:           loggingMiddleware(logger)(mux),
		ReadHeaderTimeout: 3 * time.Second,
	}
	return s
}

// func rootHandler(w http.ResponseWriter, _ *http.Request) {
// 	w.WriteHeader(http.StatusOK)
// 	_, _ = w.Write([]byte("Hi!"))
// }

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
