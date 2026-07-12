package rmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	Body []byte
	Ack  func() error
	Nack func(requeue bool) error
}

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	exchange   string
	queue      string
	routingKey string
}

func Dial(dsn, exchange, queue, routingKey string) (*Client, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	c := &Client{conn: conn, channel: ch, exchange: exchange, queue: queue, routingKey: routingKey}
	if err := c.declareTopology(); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

func (c *Client) declareTopology() error {
	if err := c.channel.ExchangeDeclare(c.exchange, amqp.ExchangeDirect, true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := c.channel.QueueDeclare(c.queue, true, false, false, false, nil); err != nil {
		return err
	}
	return c.channel.QueueBind(c.queue, c.routingKey, c.exchange, false, nil)
}

func (c *Client) Publish(ctx context.Context, body []byte) error {
	return c.channel.PublishWithContext(ctx, c.exchange, c.routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (c *Client) Consume(ctx context.Context) (<-chan Message, error) {
	deliveries, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	out := make(chan Message)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				msg := Message{
					Body: d.Body,
					Ack:  func() error { return d.Ack(false) },
					Nack: func(requeue bool) error { return d.Nack(false, requeue) },
				}
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out, nil
}

func (c *Client) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
