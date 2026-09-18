package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	inner *slog.Logger
}

func New(level string) *Logger {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn", "warning":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l})
	return &Logger{inner: slog.New(h)}
}

func (l *Logger) Info(msg string) {
	l.inner.Info(msg)
}

func (l *Logger) Warn(msg string) {
	l.inner.Warn(msg)
}

func (l *Logger) Error(msg string) {
	l.inner.Error(msg)
}

func (l *Logger) Debug(msg string) {
	l.inner.Debug(msg)
}
