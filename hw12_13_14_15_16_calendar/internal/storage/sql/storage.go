package sqlstorage

import (
	"context"
	"database/sql"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
	_ "github.com/jackc/pgx/v5"
)

type Storage struct {
	dsn string
	db  *sql.DB
}

func New(dsn string) *Storage {
	return &Storage{dsn: dsn}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return err
	}
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	return s.db.Close()
}

func (s *Storage) Add(ctx context.Context, e storage.Event) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO events (id, title, start_at, duration, description, user_id, notify_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.Title, e.StartAt, int64(e.Duration), e.Description, e.UserID, int64(e.NotifyAt),
	)
	return err
}

func (s *Storage) Update(ctx context.Context, e storage.Event) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE events
		 SET title=$2, start_at=$3, duration=$4, description=$5, user_id=$6, notify_at=$7
		 WHERE id=$1`,
		e.ID, e.Title, e.StartAt, int64(e.Duration), e.Description, e.UserID, int64(e.NotifyAt),
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (s *Storage) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (s *Storage) ListDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return s.listRange(ctx, date, date.AddDate(0, 0, 1))
}

func (s *Storage) ListWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.listRange(ctx, start, start.AddDate(0, 0, 7))
}

func (s *Storage) ListMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.listRange(ctx, start, start.AddDate(0, 1, 0))
}

func (s *Storage) listRange(ctx context.Context, from, to time.Time) ([]storage.Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, start_at, duration, description, user_id, notify_at
		 FROM events WHERE start_at >= $1 AND start_at < $2
		 ORDER BY start_at`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var (
			e        storage.Event
			dur      int64
			notifyAt int64
		)
		if err := rows.Scan(&e.ID, &e.Title, &e.StartAt, &dur, &e.Description, &e.UserID, &notifyAt); err != nil {
			return nil, err
		}
		e.Duration = time.Duration(dur)
		e.NotifyAt = time.Duration(notifyAt)
		events = append(events, e)
	}
	return events, rows.Err()
}
