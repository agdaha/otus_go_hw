package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/app"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/config"
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := config.NewConfig(configFile)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logg := logger.New(cfg.Logger.Level)

	stor, closeStorage := buildStorage(cfg, logg)
	defer closeStorage()

	calendar := app.New(logg, stor)
	server := internalhttp.NewServer(logg, calendar, cfg.HTTP.Host, cfg.HTTP.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutCancel()
		if err := server.Stop(shutCtx); err != nil {
			logg.Error("http server stop: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("http server start: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}

func buildStorage(cfg config.Config, logg *logger.Logger) (app.Storage, func()) {
	if cfg.Storage.Type == "sql" {
		s := sqlstorage.New(cfg.DB.DSN)
		connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.Connect(connectCtx); err != nil {
			connectCancel()
			log.Fatalf("db connect: %v", err)
		}
		connectCancel()
		return s, func() {
			if err := s.Close(context.Background()); err != nil {
				logg.Error("db close: " + err.Error())
			}
		}
	}
	return memorystorage.New(), func() {}
}
