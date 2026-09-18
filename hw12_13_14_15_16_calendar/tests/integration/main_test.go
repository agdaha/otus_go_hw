//go:build integration

// Package integration holds black-box integration tests that exercise the running
// calendar / calendar_scheduler / calendar_sender stack (started via docker-compose)
// through their public interfaces (HTTP API, RabbitMQ). It is a separate package so
// that `make test` (which only scans ./internal/...) never runs it; it is meant to be
// run with `make integration-tests` against a live docker-compose environment.
package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	httpAddr         = envOrDefault("CALENDAR_HTTP_ADDR", "http://localhost:8888")
	rabbitDSN        = envOrDefault("RABBIT_DSN", "amqp://rabbit:password@localhost:5672/")
	rmqExchange      = envOrDefault("RMQ_EXCHANGE", "calendar.notifications")
	statusQueue      = envOrDefault("RMQ_STATUS_QUEUE", "notifications.status")
	statusRoutingKey = envOrDefault("RMQ_STATUS_ROUTING_KEY", "notifications.status")
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestMain(m *testing.M) {
	if err := waitForHTTP(httpAddr, 60*time.Second); err != nil {
		fmt.Println("calendar HTTP API not ready:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func waitForHTTP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, addr+"/events/day?date=2020-01-01", nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		lastErr = err
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timed out waiting for %s: %w", addr, lastErr)
}
