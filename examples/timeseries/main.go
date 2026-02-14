// Example: Time series anomaly detection
//
// IoT and monitoring systems generate continuous sensor data.
// This example shows how to efficiently query anomalous readings.
//
// Run with: go run ./examples/timeseries
package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/shaia/succincter"
)

type SensorReading struct {
	Timestamp time.Time
	SensorID  string
	Value     float64
	Unit      string
}

func main() {
	fmt.Println("=== Time Series Anomaly Detection Example ===")

	// Simulate 24 hours of sensor data at 1-second intervals
	numReadings := 86400 // 24 * 60 * 60
	fmt.Printf("\nGenerating %d sensor readings (24 hours @ 1/sec)...\n", numReadings)
	readings := generateSensorData(numReadings)

	// Define anomaly thresholds
	lowThreshold := 20.0  // Below 20°C is cold anomaly
	highThreshold := 80.0 // Above 80°C is hot anomaly

	// Build indices for different anomaly types
	fmt.Println("Building anomaly indices...")
	start := time.Now()

	coldIndex := succincter.NewSuccincter(readings, func(r SensorReading) bool {
		return r.Value < lowThreshold
	})
	hotIndex := succincter.NewSuccincter(readings, func(r SensorReading) bool {
		return r.Value > highThreshold
	})
	anyAnomalyIndex := succincter.NewSuccincter(readings, func(r SensorReading) bool {
		return r.Value < lowThreshold || r.Value > highThreshold
	})

	fmt.Printf("Indices built in %v\n\n", time.Since(start))

	// Summary statistics
	totalCold := coldIndex.Rank(numReadings)
	totalHot := hotIndex.Rank(numReadings)
	totalAnomalies := anyAnomalyIndex.Rank(numReadings)

	fmt.Println("--- Anomaly Summary ---")
	fmt.Printf("Cold anomalies (<%.0f°C):  %d (%.2f%%)\n",
		lowThreshold, totalCold, float64(totalCold)*100/float64(numReadings))
	fmt.Printf("Hot anomalies (>%.0f°C):   %d (%.2f%%)\n",
		highThreshold, totalHot, float64(totalHot)*100/float64(numReadings))
	fmt.Printf("Total anomalies:          %d (%.2f%%)\n",
		totalAnomalies, float64(totalAnomalies)*100/float64(numReadings))

	// Find first few anomalies
	fmt.Println("\n--- First 5 Anomalies ---")
	for i := 1; i <= 5; i++ {
		pos := anyAnomalyIndex.Select(i)
		if pos == -1 {
			break
		}
		r := readings[pos]
		anomalyType := "COLD"
		if r.Value > highThreshold {
			anomalyType = "HOT"
		}
		fmt.Printf("  #%d: %s at %s - %.1f%s [%s]\n",
			i, r.SensorID, r.Timestamp.Format("15:04:05"), r.Value, r.Unit, anomalyType)
	}

	// Analyze anomalies by time period (hourly)
	fmt.Println("\n--- Hourly Anomaly Distribution ---")
	readingsPerHour := 3600
	for hour := 0; hour < 24; hour += 4 {
		startIdx := hour * readingsPerHour
		endIdx := (hour + 4) * readingsPerHour
		anomalies := anyAnomalyIndex.Rank(endIdx) - anyAnomalyIndex.Rank(startIdx)
		bar := makeBar(anomalies, 50)
		fmt.Printf("  %02d:00-%02d:00: %s %d\n", hour, hour+4, bar, anomalies)
	}

	// Jump to specific anomaly (pagination)
	fmt.Println("\n--- Anomaly Pagination ---")
	pageSize := 10
	page := 5
	startRank := (page-1)*pageSize + 1
	fmt.Printf("Page %d (anomalies %d-%d):\n", page, startRank, startRank+pageSize-1)
	for i := range pageSize {
		pos := anyAnomalyIndex.Select(startRank + i)
		if pos == -1 {
			break
		}
		r := readings[pos]
		fmt.Printf("  %d. Position %d: %.1f%s at %s\n",
			startRank+i, pos, r.Value, r.Unit, r.Timestamp.Format("15:04:05"))
	}

	// Time-based query: anomalies in a specific hour
	fmt.Println("\n--- Anomalies Between 14:00-15:00 ---")
	hour14Start := 14 * readingsPerHour
	hour14End := 15 * readingsPerHour
	anomaliesInHour := anyAnomalyIndex.Rank(hour14End) - anyAnomalyIndex.Rank(hour14Start)
	fmt.Printf("Count: %d anomalies\n", anomaliesInHour)
}

func generateSensorData(n int) []SensorReading {
	readings := make([]SensorReading, n)
	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Simulate temperature with daily cycle and random anomalies
	for i := range readings {
		hour := float64(i) / 3600.0
		// Base temperature: sine wave for daily cycle (40-60°C normal range)
		baseTemp := 50.0 + 10.0*math.Sin(2*math.Pi*hour/24.0)
		// Add noise
		noise := rand.NormFloat64() * 5.0
		temp := baseTemp + noise

		// Inject random anomalies (2% chance)
		if rand.Float64() < 0.02 {
			if rand.Float64() < 0.5 {
				temp = rand.Float64()*15 + 5 // Cold: 5-20°C
			} else {
				temp = rand.Float64()*20 + 80 // Hot: 80-100°C
			}
		}

		readings[i] = SensorReading{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			SensorID:  "TEMP-001",
			Value:     temp,
			Unit:      "°C",
		}
	}
	return readings
}

func makeBar(count, maxWidth int) string {
	width := min(count*maxWidth/500, maxWidth) // Scale to max 500 anomalies per period
	var bar strings.Builder
	for range width {
		bar.WriteString("█")
	}
	return bar.String()
}
