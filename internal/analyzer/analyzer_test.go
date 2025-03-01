package analyzer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/interview/junior-go-challenge/internal/models"
)

func TestLogAnalyzerProcess(t *testing.T) {
	analyzer := NewLogAnalyzer()

	// Create test entries
	entries := []models.LogEntry{
		{
			ID:        "1",
			Timestamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			Level:     models.INFO,
			Service:   "api",
			Message:   "Test message 1",
			Source:    "file1.json",
		},
		{
			ID:        "2",
			Timestamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			Level:     models.ERROR,
			Service:   "db",
			Message:   "Test message 2",
			Source:    "file1.json",
		},
	}

	// Process entries
	for _, entry := range entries {
		analyzer.Process(entry)
	}

	// Get the summary
	summary := analyzer.GetSummary()

	// Check the results
	if summary.TotalEntries != 2 {
		t.Errorf("Expected total entries to be 2, got %d", summary.TotalEntries)
	}

	if summary.ByLevel[models.INFO] != 1 {
		t.Errorf("Expected INFO level count to be 1, got %d", summary.ByLevel[models.INFO])
	}

	if summary.ByLevel[models.ERROR] != 1 {
		t.Errorf("Expected ERROR level count to be 1, got %d", summary.ByLevel[models.ERROR])
	}

	if summary.ByService["api"] != 1 {
		t.Errorf("Expected api service count to be 1, got %d", summary.ByService["api"])
	}

	if summary.ByService["db"] != 1 {
		t.Errorf("Expected db service count to be 1, got %d", summary.ByService["db"])
	}
}

func TestLogAnalyzerProcessBatch(t *testing.T) {
	analyzer := NewLogAnalyzer()

	// Create test entries
	entries := []models.LogEntry{
		{
			ID:        "1",
			Timestamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			Level:     models.INFO,
			Service:   "api",
			Message:   "Test message 1",
			Source:    "file1.json",
		},
		{
			ID:        "2",
			Timestamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			Level:     models.ERROR,
			Service:   "db",
			Message:   "Test message 2",
			Source:    "file1.json",
		},
		{
			ID:        "3",
			Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Level:     models.WARNING,
			Service:   "api",
			Message:   "Test message 3",
			Source:    "file2.json",
		},
	}

	// Process entries in batch
	analyzer.ProcessBatch(entries)

	// Wait for processing to complete
	time.Sleep(100 * time.Millisecond)

	// Get the summary
	summary := analyzer.GetSummary()

	// This test will likely fail due to concurrency issues in ProcessBatch
	if summary.TotalEntries != 3 {
		t.Errorf("Expected total entries to be 3, got %d", summary.TotalEntries)
	}
}

func TestLogAnalyzerConcurrentProcessing(t *testing.T) {
	analyzer := NewLogAnalyzer()

	// Create test entries
	entries := make([]models.LogEntry, 100)
	for i := 0; i < 100; i++ {
		entries[i] = models.LogEntry{
			ID:        fmt.Sprintf("%d", i+1),
			Timestamp: time.Date(2023, 1, 1, 10, i, 0, 0, time.UTC),
			Level:     models.INFO,
			Service:   "api",
			Message:   fmt.Sprintf("Test message %d", i+1),
			Source:    "file1.json",
		}
	}

	// Process entries concurrently
	var wg sync.WaitGroup
	for _, entry := range entries {
		wg.Add(1)
		go func(e models.LogEntry) {
			defer wg.Done()
			analyzer.Process(e)
		}(entry)
	}
	wg.Wait()

	// Get the summary
	summary := analyzer.GetSummary()

	// This test will likely fail due to race conditions in Process
	if summary.TotalEntries != 100 {
		t.Errorf("Expected total entries to be 100, got %d", summary.TotalEntries)
	}
}