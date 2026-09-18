package internalgrpc

import (
	"errors"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/api"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var errEventRequired = errors.New("event is required")

func eventFromProto(e *api.Event) (storage.Event, error) {
	if e == nil {
		return storage.Event{}, errEventRequired
	}
	return storage.Event{
		ID:          e.GetId(),
		Title:       e.GetTitle(),
		StartAt:     e.GetStartAt().AsTime(),
		Duration:    e.GetDuration().AsDuration(),
		Description: e.GetDescription(),
		UserID:      e.GetUserId(),
		NotifyAt:    e.GetNotifyAt().AsDuration(),
	}, nil
}

func eventToProto(e storage.Event) *api.Event {
	return &api.Event{
		Id:          e.ID,
		Title:       e.Title,
		StartAt:     timestamppb.New(e.StartAt),
		Duration:    durationpb.New(e.Duration),
		Description: e.Description,
		UserId:      e.UserID,
		NotifyAt:    durationpb.New(e.NotifyAt),
	}
}

func eventsToProto(events []storage.Event) []*api.Event {
	result := make([]*api.Event, 0, len(events))
	for _, e := range events {
		result = append(result, eventToProto(e))
	}
	return result
}
