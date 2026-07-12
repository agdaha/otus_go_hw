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
