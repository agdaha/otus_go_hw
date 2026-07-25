//go:build integration

package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"
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

var httpClient = &http.Client{Timeout: 5 * time.Second}

func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}

func doJSON(method, path string, payload any) (int, []byte, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, httpAddr+path, body)
	if err != nil {
		return 0, nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, respBody, nil
}

func createEvent(ev eventDTO) (int, []byte, error) {
	return doJSON(http.MethodPost, "/events", ev)
}

func deleteEvent(id string) (int, []byte, error) {
	return doJSON(http.MethodDelete, "/events/"+id, nil)
}

func listEvents(period, date string) (int, eventsResponse, error) {
	status, body, err := doJSON(http.MethodGet, "/events/"+period+"?date="+date, nil)
	if err != nil {
		return status, eventsResponse{}, err
	}
	var res eventsResponse
	if status == http.StatusOK {
		if err := json.Unmarshal(body, &res); err != nil {
			return status, eventsResponse{}, err
		}
	}
	return status, res, nil
}
