package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run generate_test_data.go <numCandles> <output.csv>")
		os.Exit(1)
	}

	var numCandles int
	_, err := fmt.Sscanf(os.Args[1], "%d", &numCandles)
	if err != nil {
		panic(err)
	}
	outputPath := os.Args[2]

	file, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"timestamp", "open", "high", "low", "close", "volume"})

	// Generate data with trends
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	basePrice := 42000.0
	price := basePrice

	rand.Seed(42) // Deterministic

	for i := 0; i < numCandles; i++ {
		timestamp := startTime.Add(time.Duration(i) * time.Hour)

		// Create trends
		trend := 0.0
		if i < 100 {
			trend = 10.0 // uptrend
		} else if i < 200 {
			trend = -8.0 // downtrend
		} else if i < 350 {
			trend = 15.0 // strong uptrend
		} else {
			trend = -5.0 // consolidation
		}

		// Random walk with trend
		change := trend + (rand.Float64()-0.5)*50
		price += change

		// Generate OHLC
		open := price
		volatility := 100.0 + rand.Float64()*100
		high := open + rand.Float64()*volatility
		low := open - rand.Float64()*volatility
		close := low + rand.Float64()*(high-low)

		// Ensure OHLC validity
		if high < open {
			high = open + volatility/2
		}
		if low > open {
			low = open - volatility/2
		}
		if close > high {
			close = high
		}
		if close < low {
			close = low
		}

		// Volume with spikes during trend changes
		baseVolume := 1000.0 + rand.Float64()*500
		volumeMultiplier := 1.0

		// Volume spike at trend changes
		if i == 100 || i == 200 || i == 350 {
			volumeMultiplier = 2.5
		} else if math.Abs(change) > 30 {
			volumeMultiplier = 1.5
		}

		volume := baseVolume * volumeMultiplier

		// Write row
		writer.Write([]string{
			timestamp.Format(time.RFC3339),
			fmt.Sprintf("%.2f", open),
			fmt.Sprintf("%.2f", high),
			fmt.Sprintf("%.2f", low),
			fmt.Sprintf("%.2f", close),
			fmt.Sprintf("%.2f", volume),
		})

		// Update price for next candle
		price = close
	}

	fmt.Printf("Generated %d candles to %s\n", numCandles, outputPath)
}
