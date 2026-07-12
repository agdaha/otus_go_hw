package sender

import (
	"context"
	"testing"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/notification"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/rmq"
	"github.com/stretchr/testify/require"
)

type capturingLogger struct {
	infos  []string
	errors []string
}

func (l *capturingLogger) Info(msg string)  { l.infos = append(l.infos, msg) }
func (l *capturingLogger) Error(msg string) { l.errors = append(l.errors, msg) }

type stubConsumer struct {
	messages chan rmq.Message
	err      error
}

func (c *stubConsumer) Consume(_ context.Context) (<-chan rmq.Message, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.messages, nil
}

func TestSenderHandlesNotification(t *testing.T) {
	body, err := notification.Notification{
		EventID: "1", Title: "t", EventAt: time.Now(), UserID: "u1",
	}.Marshal()
	require.NoError(t, err)

	acked := false
	messages := make(chan rmq.Message, 1)
	messages <- rmq.Message{Body: body, Ack: func() error { acked = true; return nil }}
	close(messages)

	logger := &capturingLogger{}
	s := New(logger, &stubConsumer{messages: messages})

	require.NoError(t, s.Run(context.Background()))
	require.True(t, acked)
	require.Len(t, logger.infos, 1)
	require.Contains(t, logger.infos[0], "event=1")
}

func TestSenderInvalidPayloadNacks(t *testing.T) {
	nacked := false
	messages := make(chan rmq.Message, 1)
	messages <- rmq.Message{Body: []byte("not-json"), Nack: func(bool) error { nacked = true; return nil }}
	close(messages)

	logger := &capturingLogger{}
	s := New(logger, &stubConsumer{messages: messages})

	require.NoError(t, s.Run(context.Background()))
	require.True(t, nacked)
	require.NotEmpty(t, logger.errors)
}

func TestSenderConsumeError(t *testing.T) {
	s := New(&capturingLogger{}, &stubConsumer{err: context.DeadlineExceeded})
	require.ErrorIs(t, s.Run(context.Background()), context.DeadlineExceeded)
}

func TestSenderStopsOnContextDone(t *testing.T) {
	messages := make(chan rmq.Message)
	s := New(&capturingLogger{}, &stubConsumer{messages: messages})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	require.NoError(t, s.Run(ctx))
}
