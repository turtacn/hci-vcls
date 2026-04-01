package logger

// Logger defines the standard logging interface for hci-vcls.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	WithFields(fields ...any) Logger
	WithError(err error) Logger
}

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

//Personal.AI order the ending
