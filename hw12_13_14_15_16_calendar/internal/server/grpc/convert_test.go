package internalgrpc

import (
	"testing"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/api"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventFromProtoNil(t *testing.T) {
	_, err := eventFromProto(nil)
	require.ErrorIs(t, err, errEventRequired)
}

func TestEventProtoRoundTrip(t *testing.T) {
	e := storage.Event{
		ID:          "1",
		Title:       "t",
		StartAt:     time.Now().Truncate(time.Second).UTC(),
		Duration:    time.Hour,
		Description: "d",
		UserID:      "u1",
		NotifyAt:    10 * time.Minute,
	}

	back, err := eventFromProto(eventToProto(e))
	require.NoError(t, err)
	require.Equal(t, e, back)
}

func TestEventsToProto(t *testing.T) {
	events := []storage.Event{
		{ID: "1", StartAt: time.Now()},
		{ID: "2", StartAt: time.Now()},
	}
	result := eventsToProto(events)
	require.Len(t, result, 2)
	require.Equal(t, "1", result[0].GetId())
	require.Equal(t, "2", result[1].GetId())

	require.Empty(t, eventsToProto(nil))
}

func TestEventToProtoFields(t *testing.T) {
	now := time.Now()
	e := storage.Event{ID: "1", StartAt: now, Duration: time.Minute}
	p := eventToProto(e)
	require.Equal(t, timestamppb.New(now).AsTime(), p.GetStartAt().AsTime())
	require.Equal(t, durationpb.New(time.Minute).AsDuration(), p.GetDuration().AsDuration())
	require.IsType(t, &api.Event{}, p)
}
