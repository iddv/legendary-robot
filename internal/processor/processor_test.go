package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/interview/junior-go-challenge/internal/analyzer"
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

	// Create sample log files with many entries to test channel blocking
	file, err := os.Create(filepath.Join(tempDir, "large.json"))
	if err != nil {
		t.Fatalf("Failed to create sample log file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for i := 0; i < 10000; i++ { // Large number of entries to ensure channel blocking
		entry := models.LogEntry{
			ID:      fmt.Sprintf("id-%d", i),
			Level:   models.INFO,
			Service: "test",
			Message: "test message",
		}
		if err := encoder.Encode(entry); err != nil {
			t.Fatalf("Failed to encode entry: %v", err)
		}
	}

	// Create a log processor
	processor := NewLogProcessor(tempDir)

	// Record initial number of goroutines
	initialGoroutines := runtime.NumGoroutine()

	// Start processing in background
	processingDone := make(chan struct{})
	go func() {
		if err := processor.Start(); err != nil {
			t.Errorf("Failed to start processor: %v", err)
		}
		close(processingDone)
	}()

	// Wait a moment for processing to begin
	time.Sleep(100 * time.Millisecond)

	// Stop the processor
	processor.Stop()

	// Wait for processing to complete with timeout
	select {
	case <-processingDone:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Processor did not shut down within timeout")
	}

	// Wait a moment for goroutines to clean up
	time.Sleep(100 * time.Millisecond)

	// Check for goroutine leaks
	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines {
		t.Errorf("Goroutine leak detected: started with %d, ended with %d", 
			initialGoroutines, finalGoroutines)
	}
}

func TestProcessorChannelBlocking(t *testing.T) {
	// Create a temporary directory for sample data
	tempDir, err := os.MkdirTemp("", "log-processor-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a large log file
	file, err := os.Create(filepath.Join(tempDir, "blocking.json"))
	if err != nil {
		t.Fatalf("Failed to create sample log file: %v", err)
	}

	// Create many log entries to test channel blocking
	encoder := json.NewEncoder(file)
	for i := 0; i < 5000; i++ {
		entry := models.LogEntry{
			ID:      fmt.Sprintf("block-%d", i),
			Level:   models.INFO,
			Service: "test",
			Message: "test message",
		}
		if err := encoder.Encode(entry); err != nil {
			t.Fatalf("Failed to encode entry: %v", err)
		}
	}
	file.Close()

	// Create a processor with a small channel buffer
	processor := &LogProcessor{
		analyzer:     analyzer.NewLogAnalyzer(),
		inputDir:     tempDir,
		batchSize:    10,
		processingCh: make(chan models.LogEntry, 10), // Small buffer to force blocking
		done:         make(chan struct{}),
	}

	// Start processing with timeout
	done := make(chan struct{})
	go func() {
		if err := processor.Start(); err != nil {
			t.Errorf("Failed to start processor: %v", err)
		}
		close(done)
	}()

	// Wait for processing with timeout
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Error("Processor blocked and did not complete within timeout")
	}
}

func TestProcessorConcurrentFiles(t *testing.T) {
	// Create a temporary directory for sample data
	tempDir, err := os.MkdirTemp("", "log-processor-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple log files
	numFiles := 10
	entriesPerFile := 100
	expectedTotal := numFiles * entriesPerFile

	for i := 0; i < numFiles; i++ {
		file, err := os.Create(filepath.Join(tempDir, fmt.Sprintf("file%d.json", i)))
		if err != nil {
			t.Fatalf("Failed to create sample log file: %v", err)
		}

		encoder := json.NewEncoder(file)
		for j := 0; j < entriesPerFile; j++ {
			entry := models.LogEntry{
				ID:      fmt.Sprintf("file%d-entry%d", i, j),
				Level:   models.INFO,
				Service: "test",
				Message: "test message",
			}
			if err := encoder.Encode(entry); err != nil {
				t.Fatalf("Failed to encode entry: %v", err)
			}
		}
		file.Close()
	}

	// Create and start processor
	processor := NewLogProcessor(tempDir)
	
	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}

	// Get summary and verify counts
	summary := processor.GetSummary()
	if summary.TotalEntries != expectedTotal {
		t.Errorf("Expected %d total entries, got %d", expectedTotal, summary.TotalEntries)
	}
}

func TestProcessorWorkerPanic(t *testing.T) {
	// Create a temporary directory for sample data
	tempDir, err := os.MkdirTemp("", "log-processor-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a log file
	file, err := os.Create(filepath.Join(tempDir, "panic.json"))
	if err != nil {
		t.Fatalf("Failed to create sample log file: %v", err)
	}

	// Create some log entries
	encoder := json.NewEncoder(file)
	for i := 0; i < 100; i++ {
		entry := models.LogEntry{
			ID:      fmt.Sprintf("panic-%d", i),
			Level:   models.INFO,
			Service: "test",
			Message: "test message",
		}
		if err := encoder.Encode(entry); err != nil {
			t.Fatalf("Failed to encode entry: %v", err)
		}
	}
	file.Close()

	// Create processor with a worker that might panic
	processor := NewLogProcessor(tempDir)
	
	// Start processing
	err = processor.Start()
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}

	// Verify the processor continues working even if a worker panics
	summary := processor.GetSummary()
	if summary.TotalEntries == 0 {
		t.Error("No entries were processed")
	}
}