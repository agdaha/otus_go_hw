package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"
)

type eventDTO struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	StartAt     time.Time `json:"startAt"`
	Duration    string    `json:"duration"`
	Description string    `json:"description,omitempty"`
	UserID      string    `json:"userId"`
	NotifyAt    string    `json:"notifyAt,omitempty"`
}

type eventsResponse struct {
	Events []eventDTO `json:"events"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	event, err := decodeEvent(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.app.CreateEvent(r.Context(), event); err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	event, err := decodeEvent(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.app.UpdateEvent(r.Context(), event); err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}
	if err := s.app.DeleteEvent(r.Context(), id); err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleListDayEvents(w http.ResponseWriter, r *http.Request) {
	handleList(w, r, s.app.ListDayEvents)
}

func (s *Server) handleListWeekEvents(w http.ResponseWriter, r *http.Request) {
	handleList(w, r, s.app.ListWeekEvents)
}

func (s *Server) handleListMonthEvents(w http.ResponseWriter, r *http.Request) {
	handleList(w, r, s.app.ListMonthEvents)
}

func handleList(
	w http.ResponseWriter,
	r *http.Request,
	list func(ctx context.Context, date time.Time) ([]storage.Event, error),
) {
	date, err := parseDate(r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	events, err := list(r.Context(), date)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, eventsResponse{Events: eventsToDTO(events)})
}

func parseDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, errors.New("date query parameter is required")
	}
	return time.Parse("2006-01-02", raw)
}

func decodeEvent(r *http.Request) (storage.Event, error) {
	var dto eventDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return storage.Event{}, err
	}
	return dto.toStorage()
}

func (d eventDTO) toStorage() (storage.Event, error) {
	dur, err := parseOptionalDuration(d.Duration)
	if err != nil {
		return storage.Event{}, err
	}
	notify, err := parseOptionalDuration(d.NotifyAt)
	if err != nil {
		return storage.Event{}, err
	}
	return storage.Event{
		ID:          d.ID,
		Title:       d.Title,
		StartAt:     d.StartAt,
		Duration:    dur,
		Description: d.Description,
		UserID:      d.UserID,
		NotifyAt:    notify,
	}, nil
}

func parseOptionalDuration(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	return time.ParseDuration(raw)
}

func eventToDTO(e storage.Event) eventDTO {
	dto := eventDTO{
		ID:          e.ID,
		Title:       e.Title,
		StartAt:     e.StartAt,
		Duration:    e.Duration.String(),
		Description: e.Description,
		UserID:      e.UserID,
	}
	if e.NotifyAt > 0 {
		dto.NotifyAt = e.NotifyAt.String()
	}
	return dto
}

func eventsToDTO(events []storage.Event) []eventDTO {
	result := make([]eventDTO, 0, len(events))
	for _, e := range events {
		result = append(result, eventToDTO(e))
	}
	return result
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		return http.StatusNotFound
	case errors.Is(err, storage.ErrDateBusy):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}
