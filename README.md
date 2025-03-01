# Log Processing Service

## Overview
This service processes log entries from multiple sources concurrently, aggregates statistics, and produces summary reports. The service simulates real-time log ingestion by processing files in a specified directory.

## Problem Description
The current implementation has concurrency issues that cause inconsistent results when processing logs. Users have reported that:

1. Sometimes log entries are processed multiple times
2. Some log entries are never processed
3. The summary statistics are inconsistent between runs with the same input data
4. The service occasionally crashes during processing
5. The performance degrades when the number of log sources increases

Your task is to identify and fix the concurrency issues in the implementation. The service should correctly process all log entries exactly once and produce consistent summary statistics.

## Setup and Running
1. Ensure you have Go installed (1.18+)
2. Clone this repository
3. Run the tests: `go test ./...`
4. Run the service: `go run cmd/logprocessor/main.go -dir ./sample-data`

## Expected Behavior
- All log entries should be processed exactly once
- The summary statistics should be consistent between runs
- The service should not crash during processing
- The performance should scale well with the number of log sources

## Testing
The project includes unit tests that verify the correct behavior of the log processor. The current implementation fails some of these tests due to concurrency issues.

Run the tests with: `go test ./...`

## Code Structure
- `cmd/logprocessor/main.go`: Entry point of the application
- `internal/processor/processor.go`: Main log processing logic
- `internal/models/log.go`: Log entry data models
- `internal/analyzer/analyzer.go`: Log analysis and statistics
- `sample-data/`: Sample log files for testing

## Hints
Look for issues related to:
- Goroutine coordination
- Shared data access
- Channel usage
- Wait group implementation
- Resource cleanup

Good luck!