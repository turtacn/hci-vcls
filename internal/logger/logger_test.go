package logger

import (
	"errors"
	"log/slog"
	"testing"
)

func TestNewLogger(t *testing.T) {
	l := NewLogger("debug", "text")
	if l == nil {
		t.Fatal("Expected logger to be created")
	}
}

func TestDefaultLogger(t *testing.T) {
	l := Default()
	if l == nil {
		t.Fatal("Expected default logger to be created")
	}
}

func TestLoggerMethods(t *testing.T) {
	// Simple test to ensure methods don't panic
	l := Default()
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")
}

func TestWithFields(t *testing.T) {
	l := Default()
	l2 := l.WithFields("key", "value")
	if l2 == nil {
		t.Fatal("Expected logger with fields to be created")
	}
	l2.Info("test message with fields")
}

func TestWithError(t *testing.T) {
	l := Default()
	l2 := l.WithError(errors.New("test error"))
	if l2 == nil {
		t.Fatal("Expected logger with error to be created")
	}
	l2.Error("test message with error")
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
	}

	for _, test := range tests {
		result := parseLevel(test.input)
		if result != test.expected {
			t.Errorf("parseLevel(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

