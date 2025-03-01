package analyzer

import (
	"sync"
	"time"

	"github.com/interview/junior-go-challenge/internal/models"
)

// LogAnalyzer aggregates statistics from log entries
type LogAnalyzer struct {
	// BUG: summary is accessed concurrently without proper synchronization
	summary    *models.LogSummary
	processedIDs map[string]bool
}

// NewLogAnalyzer creates a new log analyzer
func NewLogAnalyzer() *LogAnalyzer {
	return &LogAnalyzer{
		summary:    models.NewLogSummary(),
		processedIDs: make(map[string]bool),
	}
}

// Process analyzes a log entry and updates the summary
func (a *LogAnalyzer) Process(entry models.LogEntry) {
	// BUG: Concurrent access to maps without synchronization
	if a.processedIDs[entry.ID] {
		// Skip already processed entries
		return
	}

	// Update total count
	a.summary.TotalEntries++

	// Update counts by level
	a.summary.ByLevel[entry.Level]++

	// Update counts by service
	a.summary.ByService[entry.Service]++

	// Update time range
	if a.summary.TimeRange.Start.IsZero() || entry.Timestamp.Before(a.summary.TimeRange.Start) {
		a.summary.TimeRange.Start = entry.Timestamp
	}
	if a.summary.TimeRange.End.IsZero() || entry.Timestamp.After(a.summary.TimeRange.End) {
		a.summary.TimeRange.End = entry.Timestamp
	}

	// Mark as processed
	a.processedIDs[entry.ID] = true
}

// ProcessBatch processes multiple log entries concurrently
func (a *LogAnalyzer) ProcessBatch(entries []models.LogEntry) {
	var wg sync.WaitGroup

	// BUG: The WaitGroup is not being used correctly
	for _, entry := range entries {
		wg.Add(1)
		go func(e models.LogEntry) {
			// BUG: The following line should be deferred, but is missing
			// defer wg.Done()
			
			a.Process(e)
			
			// Simulate some processing time
			time.Sleep(time.Millisecond * 10)
		}(entry)
	}
	
	// BUG: This will likely cause a deadlock or race condition
	// Should wait for all goroutines to complete before returning
	// wg.Wait()
}

// GetSummary returns the current log summary
func (a *LogAnalyzer) GetSummary() *models.LogSummary {
	// BUG: Returns the internal data structure without making a copy
	// This allows external code to modify the internal state
	return a.summary
}