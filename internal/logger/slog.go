package logger

import (
	"log/slog"
	"os"
	"strings"

	"go.uber.org/zap"
)

type SlogLogger struct {
	l *slog.Logger
}

func NewLogger(level string, format string) Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &SlogLogger{
		l: slog.New(handler),
	}
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.l.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.l.Info(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	l.l.Warn(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	l.l.Error(msg, args...)
}

func (l *SlogLogger) WithFields(fields ...any) Logger {
	return &SlogLogger{
		l: l.l.With(fields...),
	}
}

func (l *SlogLogger) WithError(err error) Logger {
	return &SlogLogger{
		l: l.l.With("error", err),
	}
}

func (l *SlogLogger) Underlying() *zap.Logger {
	return nil
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case LevelDebug:
		return slog.LevelDebug
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Personal.AI order the ending
