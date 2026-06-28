package app

import (
	"context"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
)

type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Debug(msg string)
}

type Storage interface {
	Add(ctx context.Context, event storage.Event) error
	Update(ctx context.Context, event storage.Event) error
	Delete(ctx context.Context, id string) error
	ListDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListWeek(ctx context.Context, start time.Time) ([]storage.Event, error)
	ListMonth(ctx context.Context, start time.Time) ([]storage.Event, error)
}

type App struct {
	logger  Logger
	storage Storage
}

func New(logger Logger, storage Storage) *App {
	return &App{logger: logger, storage: storage}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	if err := a.storage.Add(ctx, event); err != nil {
		return err
	}
	a.logger.Info("event created: " + event.ID)
	return nil
}

func (a *App) UpdateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.Update(ctx, event)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	return a.storage.Delete(ctx, id)
}

func (a *App) ListDayEvents(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.ListDay(ctx, date)
}

func (a *App) ListWeekEvents(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return a.storage.ListWeek(ctx, start)
}

func (a *App) ListMonthEvents(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return a.storage.ListMonth(ctx, start)
}
