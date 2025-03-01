package models

import (
	"fmt"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	DEBUG   LogLevel = "DEBUG"
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
	FATAL   LogLevel = "FATAL"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Service   string    `json:"service"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// String returns a string representation of a LogEntry
func (l LogEntry) String() string {
	return fmt.Sprintf("%s [%s] %s: %s (source: %s)",
		l.Timestamp.Format(time.RFC3339),
		l.Level,
		l.Service,
		l.Message,
		l.Source)
}

// LogSummary contains aggregated statistics for log entries
type LogSummary struct {
	TotalEntries int
	ByLevel      map[LogLevel]int
	ByService    map[string]int
	TimeRange    struct {
		Start time.Time
		End   time.Time
	}
}

// NewLogSummary creates a new initialized LogSummary
func NewLogSummary() *LogSummary {
	return &LogSummary{
		ByLevel:   make(map[LogLevel]int),
		ByService: make(map[string]int),
	}
}
