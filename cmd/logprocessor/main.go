package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/interview/junior-go-challenge/internal/processor"
)

func main() {
	// Parse command line flags
	inputDir := flag.String("dir", "./sample-data", "Directory containing log files")
	flag.Parse()

	// Create the processor
	proc := processor.NewLogProcessor(*inputDir)

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the processor
	fmt.Println("Starting log processor...")
	err := proc.Start()
	if err != nil {
		fmt.Printf("Error starting processor: %v\n", err)
		os.Exit(1)
	}

	// Print the summary
	summary := proc.GetSummary()
	fmt.Println("\nLog Processing Summary:")
	fmt.Printf("Total Entries: %d\n", summary.TotalEntries)
	
	fmt.Println("\nEntries by Level:")
	for level, count := range summary.ByLevel {
		fmt.Printf("  %s: %d\n", level, count)
	}
	
	fmt.Println("\nEntries by Service:")
	for service, count := range summary.ByService {
		fmt.Printf("  %s: %d\n", service, count)
	}
	
	if !summary.TimeRange.Start.IsZero() && !summary.TimeRange.End.IsZero() {
		fmt.Printf("\nTime Range: %s to %s\n", 
			summary.TimeRange.Start.Format("2006-01-02 15:04:05"),
			summary.TimeRange.End.Format("2006-01-02 15:04:05"))
	}

	// Wait for signals
	select {
	case <-sigCh:
		fmt.Println("\nShutting down...")
		proc.Stop()
	}
}