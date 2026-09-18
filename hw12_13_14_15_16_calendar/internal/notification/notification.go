package notification

import (
	"encoding/json"
	"time"
)

type Notification struct {
	EventID string    `json:"eventId"`
	Title   string    `json:"title"`
	EventAt time.Time `json:"eventAt"`
	UserID  string    `json:"userId"`
}

func (n Notification) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

func Unmarshal(data []byte) (Notification, error) {
	var n Notification
	err := json.Unmarshal(data, &n)
	return n, err
}

const (
	StatusSent   = "sent"
	StatusFailed = "failed"
)

// Status is the delivery outcome the sender reports back after handling a Notification.
type Status struct {
	EventID string    `json:"eventId"`
	UserID  string    `json:"userId"`
	Status  string    `json:"status"`
	Reason  string    `json:"reason,omitempty"`
	At      time.Time `json:"at"`
}

func (s Status) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

func UnmarshalStatus(data []byte) (Status, error) {
	var s Status
	err := json.Unmarshal(data, &s)
	return s, err
}
