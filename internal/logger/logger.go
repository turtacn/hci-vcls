package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level constants
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// Logger interface abstracts the logging implementation.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	WithFields(fields ...any) Logger
	WithError(err error) Logger

	// Underlying returns the actual *zap.Logger (or nil if it's slog)
	Underlying() *zap.Logger
}

// ZapLogger is a wrapper around zap.Logger.
type ZapLogger struct {
	l *zap.Logger
}

// NewZapLogger creates a new Logger with the specified level and format.
func NewZapLogger(level string, format string) (Logger, error) {
	var config zap.Config
	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	var l zapcore.Level
	err := l.UnmarshalText([]byte(level))
	if err != nil {
		l = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(l)

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		l: zapLogger,
	}, nil
}

func (z *ZapLogger) Debug(msg string, args ...any) {
	z.l.Sugar().Debugw(msg, args...)
}

func (z *ZapLogger) Info(msg string, args ...any) {
	z.l.Sugar().Infow(msg, args...)
}

func (z *ZapLogger) Warn(msg string, args ...any) {
	z.l.Sugar().Warnw(msg, args...)
}

func (z *ZapLogger) Error(msg string, args ...any) {
	z.l.Sugar().Errorw(msg, args...)
}

func (z *ZapLogger) WithFields(fields ...any) Logger {
	return &ZapLogger{
		l: z.l.With(toZapFields(fields)...),
	}
}

func (z *ZapLogger) WithError(err error) Logger {
	return &ZapLogger{
		l: z.l.With(zap.Error(err)),
	}
}

func (z *ZapLogger) Underlying() *zap.Logger {
	return z.l
}

// Default returns a no-op logger for fallback
func Default() Logger {
	z, _ := NewZapLogger(LevelInfo, "text")
	return z
}

func toZapFields(args []any) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				fields = append(fields, zap.Any(key, args[i+1]))
			}
		}
	}
	return fields
}

