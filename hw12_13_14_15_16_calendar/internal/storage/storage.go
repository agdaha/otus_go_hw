package storage

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDateBusy      = errors.New("date is already busy")
	ErrEventNotFound = errors.New("event not found")
)

type Storage interface {
	Add(ctx context.Context, event Event) error
	Update(ctx context.Context, event Event) error
	Delete(ctx context.Context, id string) error
	ListDay(ctx context.Context, date time.Time) ([]Event, error)
	ListWeek(ctx context.Context, start time.Time) ([]Event, error)
	ListMonth(ctx context.Context, start time.Time) ([]Event, error)
}
