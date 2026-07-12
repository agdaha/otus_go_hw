// Package scheduler periodically scans the storage for events that need a notification
// and removes events that happened long ago.
package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/notification"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storage interface {
	EventsToNotify(ctx context.Context, now time.Time) ([]storage.Event, error)
	MarkNotified(ctx context.Context, id string) error
	DeleteOldEvents(ctx context.Context, before time.Time) (int64, error)
}

type Publisher interface {
	Publish(ctx context.Context, body []byte) error
}

type Scheduler struct {
	logger       Logger
	storage      Storage
	publisher    Publisher
	scanInterval time.Duration
	retention    time.Duration
}

func New(logger Logger, stor Storage, publisher Publisher, scanInterval, retention time.Duration) *Scheduler {
	return &Scheduler{
		logger:       logger,
		storage:      stor,
		publisher:    publisher,
		scanInterval: scanInterval,
		retention:    retention,
	}
}

// Run blocks, ticking every scanInterval until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	s.notify(ctx)
	s.cleanup(ctx)
}

func (s *Scheduler) notify(ctx context.Context) {
	events, err := s.storage.EventsToNotify(ctx, time.Now())
	if err != nil {
		s.logger.Error("list events to notify: " + err.Error())
		return
	}

	for _, e := range events {
		body, err := notification.Notification{
			EventID: e.ID,
			Title:   e.Title,
			EventAt: e.StartAt,
			UserID:  e.UserID,
		}.Marshal()
		if err != nil {
			s.logger.Error("marshal notification: " + err.Error())
			continue
		}

		if err := s.publisher.Publish(ctx, body); err != nil {
			s.logger.Error(fmt.Sprintf("publish notification for event %s: %v", e.ID, err))
			continue
		}

		if err := s.storage.MarkNotified(ctx, e.ID); err != nil {
			s.logger.Error(fmt.Sprintf("mark notified for event %s: %v", e.ID, err))
			continue
		}

		s.logger.Info("notification sent for event " + e.ID)
	}
}

func (s *Scheduler) cleanup(ctx context.Context) {
	n, err := s.storage.DeleteOldEvents(ctx, time.Now().Add(-s.retention))
	if err != nil {
		s.logger.Error("delete old events: " + err.Error())
		return
	}
	if n > 0 {
		s.logger.Info(fmt.Sprintf("deleted %d old events", n))
	}
}
