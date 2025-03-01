package processor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/interview/junior-go-challenge/internal/analyzer"
	"github.com/interview/junior-go-challenge/internal/models"
)

// LogProcessor processes log files and aggregates statistics
type LogProcessor struct {
	analyzer     *analyzer.LogAnalyzer
	inputDir     string
	batchSize    int
	processingCh chan models.LogEntry
	// BUG: The done channel is closed but never used properly
	done chan struct{}
}

// NewLogProcessor creates a new log processor
func NewLogProcessor(inputDir string) *LogProcessor {
	return &LogProcessor{
		analyzer:     analyzer.NewLogAnalyzer(),
		inputDir:     inputDir,
		batchSize:    100,
		processingCh: make(chan models.LogEntry, 1000),
		done:         make(chan struct{}),
	}
}

// Start begins processing log files
func (p *LogProcessor) Start() error {
	files, err := filepath.Glob(filepath.Join(p.inputDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to find log files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no log files found in directory: %s", p.inputDir)
	}

	var wg sync.WaitGroup

	// Start the workers to process log entries
	// BUG: No tracking of these workers, might lead to goroutine leaks
	for i := 0; i < 5; i++ {
		go p.worker()
	}

	// TODO

	// Process each file
	for _, file := range files {
		// BUG: Capturing loop variable in goroutine
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			err := p.processFile(file)
			if err != nil {
				fmt.Printf("Error processing file %s: %v\n", file, err)
			}
		}(file)
	}

	wg.Wait()

	// BUG: Channel is never closed, leading to goroutine leaks
	// Should close the processing channel after all files are processed
	// close(p.processingCh)

	// Simulate waiting for processing to complete
	time.Sleep(100 * time.Millisecond)

	return nil
}

// processFile reads a log file and sends entries to the processing channel
func (p *LogProcessor) processFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileName := filepath.Base(filePath)

	var entries []models.LogEntry
	decoder := json.NewDecoder(file)
	for {
		var entry models.LogEntry
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode entry: %w", err)
		}

		// Set the source to the filename
		entry.Source = fileName
		entries = append(entries, entry)
	}

	// Process entries in batches
	for i := 0; i < len(entries); i += p.batchSize {
		end := i + p.batchSize
		if end > len(entries) {
			end = len(entries)
		}
		batch := entries[i:end]

		// Send each entry to the processing channel
		for _, entry := range batch {
			// BUG: No check if the channel is closed
			// BUG: Doesn't handle blocking when the channel is full
			p.processingCh <- entry
		}
	}

	return nil
}

// worker processes log entries from the processing channel
func (p *LogProcessor) worker() {
	// BUG: No graceful shutdown mechanism
	for entry := range p.processingCh {
		// Process the entry
		p.analyzer.Process(entry)
	}
}

// GetSummary returns the current log summary
func (p *LogProcessor) GetSummary() *models.LogSummary {
	return p.analyzer.GetSummary()
}

// Stop gracefully stops the processor
func (p *LogProcessor) Stop() {
	// BUG: Closing an already closed channel will panic
	close(p.done)
}
