package internalgrpc

import (
	"context"
	"net"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/api"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc"
)

type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// Application is the subset of the calendar business logic the gRPC server depends on.
type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListDayEvents(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListWeekEvents(ctx context.Context, start time.Time) ([]storage.Event, error)
	ListMonthEvents(ctx context.Context, start time.Time) ([]storage.Event, error)
}

type Server struct {
	api.UnimplementedEventServiceServer

	grpcServer *grpc.Server
	logger     Logger
	app        Application
	addr       string
}

func NewServer(logger Logger, app Application, host, port string) *Server {
	s := &Server{
		logger: logger,
		app:    app,
		addr:   net.JoinHostPort(host, port),
	}
	s.grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(loggingInterceptor(logger)))
	api.RegisterEventServiceServer(s.grpcServer, s)
	return s
}

func (s *Server) Start(_ context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.logger.Info("grpc server listening on " + s.addr)
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop(_ context.Context) error {
	s.grpcServer.GracefulStop()
	return nil
}
