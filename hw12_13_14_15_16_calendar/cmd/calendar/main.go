package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/app"                          //nolint: depguard
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/config"                       //nolint: depguard
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/logger"                       //nolint: depguard
	internalgrpc "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/server/grpc"     //nolint: depguard
	internalhttp "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/server/http"     //nolint: depguard
	"github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage"                      //nolint: depguard
	memorystorage "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage/memory" //nolint: depguard
	sqlstorage "github.com/agdaha/otus_go_hw/hw12_13_14_15_calendar/internal/storage/sql"       //nolint: depguard
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
	httpServer := internalhttp.NewServer(logg, calendar, cfg.HTTP.Host, cfg.HTTP.Port)
	grpcServer := internalgrpc.NewServer(logg, calendar, cfg.GRPC.Host, cfg.GRPC.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutCancel()
		if err := httpServer.Stop(shutCtx); err != nil {
			logg.Error("http server stop: " + err.Error())
		}
		if err := grpcServer.Stop(shutCtx); err != nil {
			logg.Error("grpc server stop: " + err.Error())
		}
	}()

	var wg sync.WaitGroup
	failed := false

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.Start(ctx); err != nil {
			logg.Error("http server start: " + err.Error())
			failed = true
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Start(ctx); err != nil {
			logg.Error("grpc server start: " + err.Error())
			failed = true
			cancel()
		}
	}()

	logg.Info("calendar is running...")
	wg.Wait()

	if failed {
		os.Exit(1) //nolint:gocritic
	}
}

func buildStorage(cfg config.Config, logg *logger.Logger) (storage.Storage, func()) {
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
