package domain

import "time"

// LogLevel represents the severity of a log entry.
type LogLevel string

const (
	LogLevelPanic LogLevel = "PANIC"
	LogLevelFatal LogLevel = "FATAL"
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelTrace LogLevel = "TRACE"
)

// LogMessage represents a single log entry from sing-box.
type LogMessage struct {
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
