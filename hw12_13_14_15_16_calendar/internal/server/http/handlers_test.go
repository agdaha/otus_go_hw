package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

type stubLogger struct{}

func (stubLogger) Info(string)  {}
func (stubLogger) Warn(string)  {}
func (stubLogger) Error(string) {}

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

func newTestServer(app *stubApp) *Server {
	return NewServer(stubLogger{}, app, "127.0.0.1", "0")
}

func doRequest(t *testing.T, s *Server, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		req = httptest.NewRequest(method, target, bytes.NewReader(data))
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, req)
	return rec
}

func TestHandleCreateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := &stubApp{}
		s := newTestServer(app)
		rec := doRequest(t, s, http.MethodPost, "/events", eventDTO{ID: "1", Title: "t", UserID: "u1", Duration: "1h"})
		require.Equal(t, http.StatusCreated, rec.Code)
		require.Equal(t, "1", app.lastCreate.ID)
		require.Equal(t, time.Hour, app.lastCreate.Duration)
	})

	t.Run("invalid json", func(t *testing.T) {
		app := &stubApp{}
		s := newTestServer(app)
		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString("{invalid"))
		rec := httptest.NewRecorder()
		s.server.Handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid duration", func(t *testing.T) {
		app := &stubApp{}
		s := newTestServer(app)
		rec := doRequest(t, s, http.MethodPost, "/events", eventDTO{ID: "1", Duration: "not-a-duration"})
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("date busy", func(t *testing.T) {
		app := &stubApp{createErr: storage.ErrDateBusy}
		s := newTestServer(app)
		rec := doRequest(t, s, http.MethodPost, "/events", eventDTO{ID: "1"})
		require.Equal(t, http.StatusConflict, rec.Code)
	})
}

func TestHandleUpdateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := &stubApp{}
		s := newTestServer(app)
		rec := doRequest(t, s, http.MethodPut, "/events", eventDTO{ID: "1", Title: "updated"})
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "updated", app.lastUpdate.Title)
	})

	t.Run("not found", func(t *testing.T) {
		app := &stubApp{updateErr: storage.ErrEventNotFound}
		s := newTestServer(app)
		rec := doRequest(t, s, http.MethodPut, "/events", eventDTO{ID: "99"})
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandleDeleteEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := &stubApp{}
		s := newTestServer(app)
		rec := doRequest(t, s, http.MethodDelete, "/events/1", nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "1", app.lastDelete)
	})

	t.Run("not found", func(t *testing.T) {
		app := &stubApp{deleteErr: storage.ErrEventNotFound}
		s := newTestServer(app)
		rec := doRequest(t, s, http.MethodDelete, "/events/99", nil)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandleListEvents(t *testing.T) {
	base := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	sample := []storage.Event{{ID: "1", Title: "t", StartAt: base, Duration: time.Hour, UserID: "u1"}}

	tests := []struct {
		name   string
		target string
	}{
		{"day", "/events/day?date=2024-01-10"},
		{"week", "/events/week?date=2024-01-10"},
		{"month", "/events/month?date=2024-01-10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &stubApp{events: sample}
			s := newTestServer(app)
			rec := doRequest(t, s, http.MethodGet, tt.target, nil)
			require.Equal(t, http.StatusOK, rec.Code)

			var resp eventsResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
			require.Len(t, resp.Events, 1)
			require.Equal(t, "1", resp.Events[0].ID)
			require.True(t, app.lastDate.Equal(base))
		})
	}

	t.Run("missing date", func(t *testing.T) {
		s := newTestServer(&stubApp{})
		rec := doRequest(t, s, http.MethodGet, "/events/day", nil)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid date", func(t *testing.T) {
		s := newTestServer(&stubApp{})
		rec := doRequest(t, s, http.MethodGet, "/events/day?date=not-a-date", nil)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestEventDTOConversion(t *testing.T) {
	e := storage.Event{
		ID:       "1",
		Title:    "t",
		StartAt:  time.Now().Truncate(time.Second).UTC(),
		Duration: time.Hour,
		UserID:   "u1",
		NotifyAt: 10 * time.Minute,
	}
	dto := eventToDTO(e)
	back, err := dto.toStorage()
	require.NoError(t, err)
	require.Equal(t, e.ID, back.ID)
	require.True(t, e.StartAt.Equal(back.StartAt))
	require.Equal(t, e.Duration, back.Duration)
	require.Equal(t, e.NotifyAt, back.NotifyAt)
}
