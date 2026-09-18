package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "warning", "error", "unknown"} {
		l := New(level)
		if l == nil {
			t.Fatalf("New(%q) returned nil", level)
		}
	}
}

func TestLogger(t *testing.T) {
	t.Helper()
	l := New("info")
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")
}
