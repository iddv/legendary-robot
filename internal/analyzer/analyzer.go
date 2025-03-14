package analyzer

import (
	"sync"
	"time"

	"github.com/interview/junior-go-challenge/internal/models"
)

// LogAnalyzer aggregates statistics from log entries
type LogAnalyzer struct {
	mu           sync.Mutex
	summary      *models.LogSummary
	processedIDs map[string]bool
}

// NewLogAnalyzer creates a new log analyzer
func NewLogAnalyzer() *LogAnalyzer {
	return &LogAnalyzer{
		summary:      models.NewLogSummary(),
		processedIDs: make(map[string]bool),
	}
}

// Process analyzes a log entry and updates the summary
func (a *LogAnalyzer) Process(entry models.LogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

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

	for _, entry := range entries {
		wg.Add(1)
		go func(e models.LogEntry) {
			defer wg.Done()
			a.Process(e)
			
			// Simulate some processing time
			time.Sleep(time.Millisecond * 10)
		}(entry)
	}
	
	wg.Wait()
}

// GetSummary returns a copy of the current log summary
func (a *LogAnalyzer) GetSummary() *models.LogSummary {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Create a deep copy of the summary
	copy := &models.LogSummary{
		TotalEntries: a.summary.TotalEntries,
		ByLevel:      make(map[models.LogLevel]int),
		ByService:    make(map[string]int),
	}

	// Copy maps
	for k, v := range a.summary.ByLevel {
		copy.ByLevel[k] = v
	}
	for k, v := range a.summary.ByService {
		copy.ByService[k] = v
	}

	// Copy time range
	copy.TimeRange.Start = a.summary.TimeRange.Start
	copy.TimeRange.End = a.summary.TimeRange.End

	return copy
}