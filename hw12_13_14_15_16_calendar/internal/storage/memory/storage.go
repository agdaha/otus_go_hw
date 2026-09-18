package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	mu       sync.RWMutex
	events   map[string]storage.Event
	notified map[string]bool
}

func New() *Storage {
	return &Storage{
		events:   make(map[string]storage.Event),
		notified: make(map[string]bool),
	}
}

func (s *Storage) Add(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.events {
		if overlaps(e, event) {
			return storage.ErrDateBusy
		}
	}
	s.events[event.ID] = event
	return nil
}

func (s *Storage) Update(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[event.ID]; !ok {
		return storage.ErrEventNotFound
	}
	for _, e := range s.events {
		if e.ID != event.ID && overlaps(e, event) {
			return storage.ErrDateBusy
		}
	}
	s.events[event.ID] = event
	return nil
}

func (s *Storage) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	delete(s.events, id)
	delete(s.notified, id)
	return nil
}

func (s *Storage) ListDay(_ context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listRange(date, date.AddDate(0, 0, 1)), nil
}

func (s *Storage) ListWeek(_ context.Context, start time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listRange(start, start.AddDate(0, 0, 7)), nil
}

func (s *Storage) ListMonth(_ context.Context, start time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listRange(start, start.AddDate(0, 1, 0)), nil
}

func (s *Storage) EventsToNotify(_ context.Context, now time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	for _, e := range s.events {
		if e.NotifyAt <= 0 || s.notified[e.ID] {
			continue
		}
		if !e.StartAt.Add(-e.NotifyAt).After(now) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *Storage) MarkNotified(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notified[id] = true
	return nil
}

func (s *Storage) DeleteOldEvents(_ context.Context, before time.Time) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var n int64
	for id, e := range s.events {
		if e.StartAt.Before(before) {
			delete(s.events, id)
			delete(s.notified, id)
			n++
		}
	}
	return n, nil
}

func (s *Storage) listRange(from, to time.Time) []storage.Event {
	result := make([]storage.Event, 0, len(s.events))
	for _, e := range s.events {
		if !e.StartAt.Before(from) && e.StartAt.Before(to) {
			result = append(result, e)
		}
	}
	return result
}

func overlaps(a, b storage.Event) bool {
	aEnd := a.StartAt.Add(a.Duration)
	bEnd := b.StartAt.Add(b.Duration)
	return a.StartAt.Before(bEnd) && b.StartAt.Before(aEnd)
}
