package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/rmq"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/sender"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/sender_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cfg, err := NewConfig(configFile)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := run(cfg); err != nil {
		log.Fatalf("sender: %v", err)
	}
}

func run(cfg Config) error {
	logg := logger.New(cfg.Logger.Level)

	client := buildRMQClient(cfg)
	defer func() {
		if err := client.Close(); err != nil {
			logg.Error("rmq close: " + err.Error())
		}
	}()

	snd := sender.New(logg, client)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg.Info("calendar_sender is running...")
	return snd.Run(ctx)
}

func buildRMQClient(cfg Config) *rmq.Client {
	client, err := rmq.Dial(cfg.RMQ.DSN, cfg.RMQ.Exchange, cfg.RMQ.Queue, cfg.RMQ.RoutingKey)
	if err != nil {
		log.Fatalf("rmq dial: %v", err)
	}
	return client
}
