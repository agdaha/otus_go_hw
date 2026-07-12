package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/rmq"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/scheduler"
	memorystorage "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage/sql"
)

const (
	defaultScanInterval = time.Minute
	defaultRetention    = 365 * 24 * time.Hour
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/scheduler_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cfg, err := NewConfig(configFile)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := run(cfg); err != nil {
		log.Fatalf("scheduler: %v", err)
	}
}

func run(cfg Config) error {
	logg := logger.New(cfg.Logger.Level)

	stor, closeStorage := buildStorage(cfg)
	client := buildRMQClient(cfg)

	defer closeStorage()
	defer func() {
		if err := client.Close(); err != nil {
			logg.Error("rmq close: " + err.Error())
		}
	}()

	scanInterval := parseDurationOrDefault(cfg.Scheduler.ScanInterval, defaultScanInterval)
	retention := parseDurationOrDefault(cfg.Scheduler.Retention, defaultRetention)

	sched := scheduler.New(logg, stor, client, scanInterval, retention)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg.Info("calendar_scheduler is running...")
	return sched.Run(ctx)
}

func buildStorage(cfg Config) (scheduler.Storage, func()) {
	if cfg.Storage.Type == "memory" {
		return memorystorage.New(), func() {}
	}

	s := sqlstorage.New(cfg.DB.DSN)
	connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := s.Connect(connectCtx); err != nil {
		connectCancel()
		log.Fatalf("db connect: %v", err)
	}
	connectCancel()
	return s, func() {
		if err := s.Close(context.Background()); err != nil {
			log.Printf("db close: %v", err)
		}
	}
}

func buildRMQClient(cfg Config) *rmq.Client {
	client, err := rmq.Dial(cfg.RMQ.DSN, cfg.RMQ.Exchange, cfg.RMQ.Queue, cfg.RMQ.RoutingKey)
	if err != nil {
		log.Fatalf("rmq dial: %v", err)
	}
	return client
}

func parseDurationOrDefault(raw string, def time.Duration) time.Duration {
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return def
	}
	return d
}
