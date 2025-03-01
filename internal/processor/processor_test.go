package processor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/interview/junior-go-challenge/internal/models"
)

func createSampleLogs(t *testing.T, dir string) {
	// Create sample data directory
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sample data directory: %v", err)
	}

	// Create log entries for file 1
	entries1 := []models.LogEntry{
		{
			ID:        "1",
			Timestamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			Level:     models.INFO,
			Service:   "api",
			Message:   "User login successful",
		},
		{
			ID:        "2",
			Timestamp: time.Date(2023, 1, 1, 10, 5, 0, 0, time.UTC),
			Level:     models.ERROR,
			Service:   "db",
			Message:   "Connection timeout",
		},
	}

	// Write log entries to file 1
	file1, err := os.Create(filepath.Join(dir, "logs1.json"))
	if err != nil {
		t.Fatalf("Failed to create sample log file: %v", err)
	}
	defer file1.Close()

	encoder1 := json.NewEncoder(file1)
	for _, entry := range entries1 {
		err := encoder1.Encode(entry)
		if err != nil {
			t.Fatalf("Failed to encode log entry: %v", err)
		}
	}

	// Create log entries for file 2
	entries2 := []models.LogEntry{
		{
			ID:        "3",
			Timestamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			Level:     models.WARNING,
			Service:   "api",
			Message:   "High memory usage",
		},
		{
			ID:        "4",
			Timestamp: time.Date(2023, 1, 1, 11, 5, 0, 0, time.UTC),
			Level:     models.INFO,
			Service:   "auth",
			Message:   "Token refreshed",
		},
		{
			ID:        "5",
			Timestamp: time.Date(2023, 1, 1, 11, 10, 0, 0, time.UTC),
			Level:     models.DEBUG,
			Service:   "api",
			Message:   "Request parameters: {...}",
		},
	}

	// Write log entries to file 2
	file2, err := os.Create(filepath.Join(dir, "logs2.json"))
	if err != nil {
		t.Fatalf("Failed to create sample log file: %v", err)
	}
	defer file2.Close()

	encoder2 := json.NewEncoder(file2)
	for _, entry := range entries2 {
		err := encoder2.Encode(entry)
		if err != nil {
			t.Fatalf("Failed to encode log entry: %v", err)
		}
	}
}

func TestProcessorStart(t *testing.T) {
	// Create a temporary directory for sample data
	tempDir, err := os.MkdirTemp("", "log-processor-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create sample log files
	createSampleLogs(t, tempDir)

	// Create a log processor
	processor := NewLogProcessor(tempDir)

	// Start the processor
	err = processor.Start()
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}

	// Wait for processing to complete
	time.Sleep(500 * time.Millisecond)

	// Get the summary
	summary := processor.GetSummary()

	// This test may fail due to concurrency issues
	if summary.TotalEntries != 5 {
		t.Errorf("Expected total entries to be 5, got %d", summary.TotalEntries)
	}

	// Check level counts
	expectedLevelCounts := map[models.LogLevel]int{
		models.DEBUG:   1,
		models.INFO:    2,
		models.WARNING: 1,
		models.ERROR:   1,
	}

	for level, expectedCount := range expectedLevelCounts {
		if summary.ByLevel[level] != expectedCount {
			t.Errorf("Expected %s level count to be %d, got %d", level, expectedCount, summary.ByLevel[level])
		}
	}

	// Check service counts
	expectedServiceCounts := map[string]int{
		"api":  3,
		"db":   1,
		"auth": 1,
	}

	for service, expectedCount := range expectedServiceCounts {
		if summary.ByService[service] != expectedCount {
			t.Errorf("Expected %s service count to be %d, got %d", service, expectedCount, summary.ByService[service])
		}
	}
}

func TestProcessorGracefulShutdown(t *testing.T) {
	// Create a temporary directory for sample data
	tempDir, err := os.MkdirTemp("", "log-processor-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create sample log files
	createSampleLogs(t, tempDir)

	// Create a log processor
	processor := NewLogProcessor(tempDir)

	// Start the processor in a goroutine
	go func() {
		err := processor.Start()
		if err != nil {
			t.Errorf("Failed to start processor: %v", err)
		}
	}()

	// Wait a moment before stopping
	time.Sleep(100 * time.Millisecond)

	// Stop the processor - this may panic due to the bug in Stop()
	// This test will likely fail
	processor.Stop()

	// Wait for processing to complete
	time.Sleep(100 * time.Millisecond)
}