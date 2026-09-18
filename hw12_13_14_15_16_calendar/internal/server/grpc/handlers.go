package internalgrpc

import (
	"context"
	"errors"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/api"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateEvent(ctx context.Context, req *api.CreateEventRequest) (*emptypb.Empty, error) {
	event, err := eventFromProto(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.app.CreateEvent(ctx, event); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *api.UpdateEventRequest) (*emptypb.Empty, error) {
	event, err := eventFromProto(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.app.UpdateEvent(ctx, event); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *api.DeleteEventRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteEvent(ctx, req.GetId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListDayEvents(ctx context.Context, req *api.ListEventsRequest) (*api.ListEventsResponse, error) {
	events, err := s.app.ListDayEvents(ctx, req.GetDate().AsTime())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &api.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func (s *Server) ListWeekEvents(ctx context.Context, req *api.ListEventsRequest) (*api.ListEventsResponse, error) {
	events, err := s.app.ListWeekEvents(ctx, req.GetDate().AsTime())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &api.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func (s *Server) ListMonthEvents(ctx context.Context, req *api.ListEventsRequest) (*api.ListEventsResponse, error) {
	events, err := s.app.ListMonthEvents(ctx, req.GetDate().AsTime())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &api.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, storage.ErrDateBusy):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
