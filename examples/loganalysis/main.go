// Example: Using Succincter for efficient log analysis
//
// Run with: go run ./examples/loganalysis
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shaia/succincter"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
}

func main() {
	fmt.Println("=== Succincter Log Analysis Example ===")

	// Generate 1 million log entries (5% are errors)
	numEntries := 1_000_000
	errorRate := 0.05
	fmt.Printf("Generating %d log entries (%.0f%% errors)...\n", numEntries, errorRate*100)

	logs := generateLogs(numEntries, errorRate)

	// Build Succincter index
	fmt.Println("Building Succincter index...")
	start := time.Now()
	errorIndex := succincter.NewSuccincter(logs, func(e LogEntry) bool {
		return e.Level == "ERROR"
	})
	buildTime := time.Since(start)
	fmt.Printf("Index built in %v\n\n", buildTime)

	// Demonstrate queries
	fmt.Println("\n--- Query Examples ---")

	// Query 1: Count errors before position
	pos := 500_000
	start = time.Now()
	count := errorIndex.Rank(pos)
	queryTime := time.Since(start)
	fmt.Printf("1. Errors before position %d: %d (query time: %v)\n", pos, count, queryTime)

	// Query 2: Find Nth error
	n := 1000
	start = time.Now()
	errorPos := errorIndex.Select(n)
	queryTime = time.Since(start)
	fmt.Printf("2. Position of error #%d: %d (query time: %v)\n", n, errorPos, queryTime)
	fmt.Printf("   Verification: logs[%d].Level = %q\n", errorPos, logs[errorPos].Level)

	// Query 3: Count errors in a range
	rangeStart, rangeEnd := 100_000, 200_000
	start = time.Now()
	errorsInRange := errorIndex.Rank(rangeEnd) - errorIndex.Rank(rangeStart)
	queryTime = time.Since(start)
	fmt.Printf("3. Errors in range [%d, %d): %d (query time: %v)\n",
		rangeStart, rangeEnd, errorsInRange, queryTime)

	// Query 4: List first 5 errors
	fmt.Println("\n4. First 5 errors:")
	for i := 1; i <= 5; i++ {
		p := errorIndex.Select(i)
		fmt.Printf("   Error %d at position %d: %s\n", i, p, logs[p].Message)
	}

	// Compare with naive approach
	fmt.Println("\n--- Performance Comparison ---")
	comparePerformance(logs, errorIndex)
}

func generateLogs(n int, errorRate float64) []LogEntry {
	levels := []string{"DEBUG", "INFO", "WARN"}
	logs := make([]LogEntry, n)

	for i := range logs {
		level := levels[rand.Intn(3)]
		if rand.Float64() < errorRate {
			level = "ERROR"
		}
		logs[i] = LogEntry{
			Timestamp: time.Now().Add(time.Duration(i) * time.Millisecond),
			Level:     level,
			Message:   fmt.Sprintf("Log message %d", i),
		}
	}
	return logs
}

func comparePerformance(logs []LogEntry, index *succincter.Succincter) {
	testPos := 500_000

	// Naive Rank (fewer iterations - it's slow)
	naiveIterations := 100
	start := time.Now()
	for i := 0; i < naiveIterations; i++ {
		naiveCountBefore(logs, testPos)
	}
	naiveTotal := time.Since(start)
	naiveAvg := naiveTotal / time.Duration(naiveIterations)

	// Succincter Rank (many more iterations - it's fast)
	succincterIterations := 1_000_000
	start = time.Now()
	for i := 0; i < succincterIterations; i++ {
		index.Rank(testPos)
	}
	succincterTotal := time.Since(start)
	succincterAvg := succincterTotal / time.Duration(succincterIterations)

	fmt.Printf("Rank(%d):\n", testPos)
	fmt.Printf("  Naive:      %v avg (%d iterations)\n", naiveAvg, naiveIterations)
	fmt.Printf("  Succincter: %v avg (%d iterations)\n", succincterAvg, succincterIterations)
	fmt.Printf("  Speedup:    %.0fx\n", float64(naiveAvg)/float64(succincterAvg))
}

func naiveCountBefore(logs []LogEntry, pos int) int {
	count := 0
	for i := 0; i < pos && i < len(logs); i++ {
		if logs[i].Level == "ERROR" {
			count++
		}
	}
	return count
}
