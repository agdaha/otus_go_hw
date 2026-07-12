// Package sender reads notifications from the queue and logs them.
package sender

import (
	"context"
	"fmt"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/notification"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/rmq"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Consumer interface {
	Consume(ctx context.Context) (<-chan rmq.Message, error)
}

type Sender struct {
	logger   Logger
	consumer Consumer
}

func New(logger Logger, consumer Consumer) *Sender {
	return &Sender{logger: logger, consumer: consumer}
}

// Run blocks, handling messages until the channel closes or ctx is cancelled.
func (s *Sender) Run(ctx context.Context) error {
	messages, err := s.consumer.Consume(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-messages:
			if !ok {
				return nil
			}
			s.handle(msg)
		}
	}
}

func (s *Sender) handle(msg rmq.Message) {
	n, err := notification.Unmarshal(msg.Body)
	if err != nil {
		s.logger.Error("decode notification: " + err.Error())
		if nackErr := msg.Nack(false); nackErr != nil {
			s.logger.Error("nack message: " + nackErr.Error())
		}
		return
	}

	s.logger.Info(fmt.Sprintf(
		"notification: event=%s title=%q at=%s user=%s",
		n.EventID, n.Title, n.EventAt.Format("2006-01-02T15:04:05Z07:00"), n.UserID,
	))

	if err := msg.Ack(); err != nil {
		s.logger.Error("ack message: " + err.Error())
	}
}
