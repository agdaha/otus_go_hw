package internalgrpc

import (
	"context"
	"testing"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/api"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type stubApp struct {
	createErr error
	updateErr error
	deleteErr error
	listErr   error
	events    []storage.Event

	lastCreate storage.Event
	lastUpdate storage.Event
	lastDelete string
	lastDate   time.Time
}

func (a *stubApp) CreateEvent(_ context.Context, e storage.Event) error {
	a.lastCreate = e
	return a.createErr
}

func (a *stubApp) UpdateEvent(_ context.Context, e storage.Event) error {
	a.lastUpdate = e
	return a.updateErr
}

func (a *stubApp) DeleteEvent(_ context.Context, id string) error {
	a.lastDelete = id
	return a.deleteErr
}

func (a *stubApp) ListDayEvents(_ context.Context, date time.Time) ([]storage.Event, error) {
	a.lastDate = date
	return a.events, a.listErr
}

func (a *stubApp) ListWeekEvents(_ context.Context, date time.Time) ([]storage.Event, error) {
	a.lastDate = date
	return a.events, a.listErr
}

func (a *stubApp) ListMonthEvents(_ context.Context, date time.Time) ([]storage.Event, error) {
	a.lastDate = date
	return a.events, a.listErr
}

func TestServerCreateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := &stubApp{}
		s := &Server{app: app}
		req := &api.CreateEventRequest{Event: &api.Event{
			Id:       "1",
			Title:    "t",
			StartAt:  timestamppb.New(time.Now()),
			Duration: durationpb.New(time.Hour),
			UserId:   "u1",
		}}

		_, err := s.CreateEvent(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, "1", app.lastCreate.ID)
		require.Equal(t, time.Hour, app.lastCreate.Duration)
	})

	t.Run("nil event", func(t *testing.T) {
		s := &Server{app: &stubApp{}}
		_, err := s.CreateEvent(context.Background(), &api.CreateEventRequest{})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("date busy", func(t *testing.T) {
		s := &Server{app: &stubApp{createErr: storage.ErrDateBusy}}
		_, err := s.CreateEvent(context.Background(), &api.CreateEventRequest{Event: &api.Event{Id: "1"}})
		require.Equal(t, codes.AlreadyExists, status.Code(err))
	})
}

func TestServerUpdateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := &stubApp{}
		s := &Server{app: app}
		req := &api.UpdateEventRequest{Event: &api.Event{Id: "1", Title: "updated"}}

		_, err := s.UpdateEvent(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, "updated", app.lastUpdate.Title)
	})

	t.Run("not found", func(t *testing.T) {
		s := &Server{app: &stubApp{updateErr: storage.ErrEventNotFound}}
		_, err := s.UpdateEvent(context.Background(), &api.UpdateEventRequest{Event: &api.Event{Id: "99"}})
		require.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("nil event", func(t *testing.T) {
		s := &Server{app: &stubApp{}}
		_, err := s.UpdateEvent(context.Background(), &api.UpdateEventRequest{})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestServerDeleteEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := &stubApp{}
		s := &Server{app: app}
		_, err := s.DeleteEvent(context.Background(), &api.DeleteEventRequest{Id: "1"})
		require.NoError(t, err)
		require.Equal(t, "1", app.lastDelete)
	})

	t.Run("not found", func(t *testing.T) {
		s := &Server{app: &stubApp{deleteErr: storage.ErrEventNotFound}}
		_, err := s.DeleteEvent(context.Background(), &api.DeleteEventRequest{Id: "99"})
		require.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestServerListEvents(t *testing.T) {
	base := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	sample := []storage.Event{{ID: "1", Title: "t", StartAt: base, Duration: time.Hour, UserID: "u1"}}

	tests := []struct {
		name string
		list func(*Server, context.Context, *api.ListEventsRequest) (*api.ListEventsResponse, error)
	}{
		{"day", (*Server).ListDayEvents},
		{"week", (*Server).ListWeekEvents},
		{"month", (*Server).ListMonthEvents},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &stubApp{events: sample}
			s := &Server{app: app}

			resp, err := tt.list(s, context.Background(), &api.ListEventsRequest{Date: timestamppb.New(base)})
			require.NoError(t, err)
			require.Len(t, resp.GetEvents(), 1)
			require.Equal(t, "1", resp.GetEvents()[0].GetId())
			require.True(t, app.lastDate.Equal(base))
		})
	}

	t.Run("internal error", func(t *testing.T) {
		s := &Server{app: &stubApp{listErr: context.DeadlineExceeded}}
		_, err := s.ListDayEvents(context.Background(), &api.ListEventsRequest{Date: timestamppb.New(base)})
		require.Equal(t, codes.Internal, status.Code(err))
	})
}
