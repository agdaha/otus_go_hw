package memorystorage

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

func makeEvent(id string, start time.Time, dur time.Duration) storage.Event {
	return storage.Event{ID: id, Title: "t", UserID: "u1", StartAt: start, Duration: dur}
}

func TestAdd(t *testing.T) {
	s := New()
	ctx := context.Background()
	e := makeEvent("1", time.Now(), time.Hour)

	require.NoError(t, s.Add(ctx, e))
	require.ErrorIs(t, s.Add(ctx, e), storage.ErrDateBusy)
}

func TestAddOverlap(t *testing.T) {
	s := New()
	ctx := context.Background()
	now := time.Now()

	require.NoError(t, s.Add(ctx, makeEvent("1", now, 2*time.Hour)))
	require.ErrorIs(t, s.Add(ctx, makeEvent("2", now.Add(time.Hour), time.Hour)), storage.ErrDateBusy)
	require.NoError(t, s.Add(ctx, makeEvent("3", now.Add(3*time.Hour), time.Hour)))
}

func TestUpdate(t *testing.T) {
	s := New()
	ctx := context.Background()
	now := time.Now()
	e := makeEvent("1", now, time.Hour)

	require.NoError(t, s.Add(ctx, e))
	e.Title = "updated"
	require.NoError(t, s.Update(ctx, e))
	require.ErrorIs(t, s.Update(ctx, makeEvent("99", now.Add(5*time.Hour), time.Hour)), storage.ErrEventNotFound)
}

func TestDelete(t *testing.T) {
	s := New()
	ctx := context.Background()

	require.NoError(t, s.Add(ctx, makeEvent("1", time.Now(), time.Hour)))
	require.NoError(t, s.Delete(ctx, "1"))
	require.ErrorIs(t, s.Delete(ctx, "1"), storage.ErrEventNotFound)
}

func TestListDay(t *testing.T) {
	s := New()
	ctx := context.Background()
	base := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)

	require.NoError(t, s.Add(ctx, makeEvent("1", base, time.Minute)))
	require.NoError(t, s.Add(ctx, makeEvent("2", base.AddDate(0, 0, 1), time.Minute)))

	events, err := s.ListDay(ctx, base)
	require.NoError(t, err)
	require.Len(t, events, 1)
}

func TestListWeek(t *testing.T) {
	s := New()
	ctx := context.Background()
	base := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)

	for i := range 7 {
		require.NoError(t, s.Add(ctx, makeEvent(strconv.Itoa(i), base.AddDate(0, 0, i), time.Minute)))
	}
	require.NoError(t, s.Add(ctx, makeEvent("x", base.AddDate(0, 0, 7), time.Minute)))

	events, err := s.ListWeek(ctx, base)
	require.NoError(t, err)
	require.Len(t, events, 7)
}

func TestListMonth(t *testing.T) {
	s := New()
	ctx := context.Background()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range 31 {
		require.NoError(t, s.Add(ctx, makeEvent(strconv.Itoa(i), base.AddDate(0, 0, i), time.Minute)))
	}

	events, err := s.ListMonth(ctx, base)
	require.NoError(t, err)
	require.Len(t, events, 31)
}

func TestEventsToNotify(t *testing.T) {
	s := New()
	ctx := context.Background()
	now := time.Now()

	due := storage.Event{ID: "due", UserID: "u1", StartAt: now.Add(time.Minute), NotifyAt: 10 * time.Minute}
	notDue := storage.Event{ID: "not-due", UserID: "u1", StartAt: now.Add(time.Hour), NotifyAt: time.Minute}
	noNotify := storage.Event{ID: "no-notify", UserID: "u1", StartAt: now}

	require.NoError(t, s.Add(ctx, due))
	require.NoError(t, s.Add(ctx, notDue))
	require.NoError(t, s.Add(ctx, noNotify))

	events, err := s.EventsToNotify(ctx, now)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "due", events[0].ID)

	require.NoError(t, s.MarkNotified(ctx, "due"))
	events, err = s.EventsToNotify(ctx, now)
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestDeleteOldEvents(t *testing.T) {
	s := New()
	ctx := context.Background()
	now := time.Now()

	require.NoError(t, s.Add(ctx, makeEvent("old", now.AddDate(-2, 0, 0), time.Hour)))
	require.NoError(t, s.Add(ctx, makeEvent("recent", now, time.Hour)))

	n, err := s.DeleteOldEvents(ctx, now.AddDate(-1, 0, 0))
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	require.ErrorIs(t, s.Delete(ctx, "old"), storage.ErrEventNotFound)
	require.NoError(t, s.Delete(ctx, "recent"))
}

func TestConcurrentAdd(_ *testing.T) {
	s := New()
	ctx := context.Background()
	base := time.Now()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			e := makeEvent(strconv.Itoa(i), base.Add(time.Duration(i)*time.Hour), time.Minute)
			_ = s.Add(ctx, e)
		}(i)
	}
	wg.Wait()
}
