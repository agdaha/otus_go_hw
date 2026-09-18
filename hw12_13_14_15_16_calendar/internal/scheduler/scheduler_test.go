package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/notification"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

type stubLogger struct{}

func (stubLogger) Info(string)  {}
func (stubLogger) Error(string) {}

type stubStorage struct {
	events        []storage.Event
	notifyErr     error
	markErr       error
	deleteErr     error
	notifiedIDs   []string
	deletedBefore time.Time
	deleteCount   int64
}

func (s *stubStorage) EventsToNotify(_ context.Context, _ time.Time) ([]storage.Event, error) {
	return s.events, s.notifyErr
}

func (s *stubStorage) MarkNotified(_ context.Context, id string) error {
	s.notifiedIDs = append(s.notifiedIDs, id)
	return s.markErr
}

func (s *stubStorage) DeleteOldEvents(_ context.Context, before time.Time) (int64, error) {
	s.deletedBefore = before
	return s.deleteCount, s.deleteErr
}

type stubPublisher struct {
	published [][]byte
	err       error
}

func (p *stubPublisher) Publish(_ context.Context, body []byte) error {
	p.published = append(p.published, body)
	return p.err
}

func TestSchedulerNotify(t *testing.T) {
	t.Run("publishes and marks notified", func(t *testing.T) {
		start := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
		stor := &stubStorage{events: []storage.Event{
			{ID: "1", Title: "t", StartAt: start, UserID: "u1", NotifyAt: time.Minute},
		}}
		pub := &stubPublisher{}
		s := New(stubLogger{}, stor, pub, time.Minute, time.Hour)

		s.notify(context.Background())

		require.Len(t, pub.published, 1)
		n, err := notification.Unmarshal(pub.published[0])
		require.NoError(t, err)
		require.Equal(t, "1", n.EventID)
		require.Equal(t, "t", n.Title)
		require.Equal(t, "u1", n.UserID)
		require.Equal(t, []string{"1"}, stor.notifiedIDs)
	})

	t.Run("publish error skips mark notified", func(t *testing.T) {
		stor := &stubStorage{events: []storage.Event{{ID: "1"}}}
		pub := &stubPublisher{err: context.DeadlineExceeded}
		s := New(stubLogger{}, stor, pub, time.Minute, time.Hour)

		s.notify(context.Background())

		require.Empty(t, stor.notifiedIDs)
	})

	t.Run("list error is a no-op", func(t *testing.T) {
		stor := &stubStorage{notifyErr: context.DeadlineExceeded}
		pub := &stubPublisher{}
		s := New(stubLogger{}, stor, pub, time.Minute, time.Hour)

		s.notify(context.Background())

		require.Empty(t, pub.published)
	})
}

func TestSchedulerCleanup(t *testing.T) {
	stor := &stubStorage{deleteCount: 3}
	s := New(stubLogger{}, stor, &stubPublisher{}, time.Minute, 365*24*time.Hour)

	before := time.Now()
	s.cleanup(context.Background())

	require.WithinDuration(t, before.Add(-365*24*time.Hour), stor.deletedBefore, time.Second)
}

func TestSchedulerRunStopsOnContextDone(t *testing.T) {
	stor := &stubStorage{}
	s := New(stubLogger{}, stor, &stubPublisher{}, time.Millisecond, time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	require.NoError(t, s.Run(ctx))
}
