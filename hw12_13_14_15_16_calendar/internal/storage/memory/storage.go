package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{events: make(map[string]storage.Event)}
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

func (s *Storage) listRange(from, to time.Time) []storage.Event {
	result := make([]storage.Event, 0, len(s.events))
	for _, e := range s.events {
		if !e.StartAt.Before(from) && e.StartAt.Before(to) {
			result = append(result, e)
		}
	}
	return result
}

// overlaps reports whether two events occupy the same time slot.
func overlaps(a, b storage.Event) bool {
	aEnd := a.StartAt.Add(a.Duration)
	bEnd := b.StartAt.Add(b.Duration)
	return a.StartAt.Before(bEnd) && b.StartAt.Before(aEnd)
}
