//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateEvent_Success(t *testing.T) {
	id := newID("create")
	defer func() { _, _, _ = deleteEvent(id) }()

	status, body, err := createEvent(eventDTO{
		ID:       id,
		Title:    "integration test event",
		StartAt:  time.Now().Add(48 * time.Hour).Truncate(time.Second),
		Duration: "1h",
		UserID:   "itest-user",
	})
	require.NoError(t, err)
	require.Equalf(t, http.StatusCreated, status, "body: %s", body)
}

func TestCreateEvent_DateBusyBusinessError(t *testing.T) {
	id1, id2 := newID("busy-1"), newID("busy-2")
	defer func() {
		_, _, _ = deleteEvent(id1)
		_, _, _ = deleteEvent(id2)
	}()

	start := time.Now().Add(72 * time.Hour).Truncate(time.Second)

	status, body, err := createEvent(eventDTO{ID: id1, Title: "first", StartAt: start, Duration: "1h", UserID: "itest-user"})
	require.NoError(t, err)
	require.Equalf(t, http.StatusCreated, status, "body: %s", body)

	status, body, err = createEvent(eventDTO{
		ID: id2, Title: "second", StartAt: start.Add(30 * time.Minute), Duration: "1h", UserID: "itest-user",
	})
	require.NoError(t, err)
	require.Equalf(t, http.StatusConflict, status, "body: %s", body)

	var errResp errorResponse
	require.NoError(t, json.Unmarshal(body, &errResp))
	require.NotEmpty(t, errResp.Error)
}

func TestCreateEvent_InvalidPayloadBusinessError(t *testing.T) {
	status, _, err := createEvent(eventDTO{ID: newID("bad"), Duration: "not-a-duration"})
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, status)
}

func TestListEvents_DayWeekMonth(t *testing.T) {
	id := newID("list")
	defer func() { _, _, _ = deleteEvent(id) }()

	day := time.Now().AddDate(0, 0, 10).Truncate(24 * time.Hour).UTC()

	status, body, err := createEvent(eventDTO{
		ID: id, Title: "listed event", StartAt: day.Add(2 * time.Hour), Duration: "1h", UserID: "itest-user",
	})
	require.NoError(t, err)
	require.Equalf(t, http.StatusCreated, status, "body: %s", body)

	dateStr := day.Format("2006-01-02")

	for _, period := range []string{"day", "week", "month"} {
		t.Run(period, func(t *testing.T) {
			status, res, err := listEvents(period, dateStr)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
			require.True(t, containsEventID(res.Events, id), "expected event %s in %s listing", id, period)
		})
	}
}

func containsEventID(events []eventDTO, id string) bool {
	for _, e := range events {
		if e.ID == id {
			return true
		}
	}
	return false
}
