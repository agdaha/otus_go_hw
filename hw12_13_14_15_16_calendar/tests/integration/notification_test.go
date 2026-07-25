//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/notification"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

// TestNotificationDelivery creates an event due for notification almost immediately and
// verifies that calendar_scheduler publishes it and calendar_sender reports a "sent"
// status back over the RabbitMQ status queue.
func TestNotificationDelivery(t *testing.T) {
	conn, err := amqp.Dial(rabbitDSN)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	require.NoError(t, ch.ExchangeDeclare(rmqExchange, amqp.ExchangeDirect, true, false, false, false, nil))
	_, err = ch.QueueDeclare(statusQueue, true, false, false, false, nil)
	require.NoError(t, err)
	require.NoError(t, ch.QueueBind(statusQueue, statusRoutingKey, rmqExchange, false, nil))

	deliveries, err := ch.Consume(statusQueue, "", true, false, false, false, nil)
	require.NoError(t, err)

	id := newID("notify")
	defer func() { _, _, _ = deleteEvent(id) }()

	status, body, err := createEvent(eventDTO{
		ID:       id,
		Title:    "notify me",
		StartAt:  time.Now().Add(3 * time.Second),
		Duration: "1h",
		UserID:   "itest-user",
		NotifyAt: "2s",
	})
	require.NoError(t, err)
	require.Equalf(t, http.StatusCreated, status, "body: %s", body)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timed out waiting for notification status for event %s", id)
		case d, ok := <-deliveries:
			if !ok {
				t.Fatal("status queue delivery channel closed")
			}
			st, err := notification.UnmarshalStatus(d.Body)
			require.NoError(t, err)
			if st.EventID != id {
				continue
			}
			require.Equal(t, notification.StatusSent, st.Status)
			return
		}
	}
}
