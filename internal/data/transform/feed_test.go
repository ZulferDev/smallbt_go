package transform

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/data/csv"
)

func createTestCSV(tb testing.TB, count int) string {
	tb.Helper()

	tmpDir := tb.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")

	file, err := os.Create(csvPath)
	if err != nil {
		tb.Fatalf("Failed to create temp CSV: %v", err)
	}
	defer file.Close()

	_, _ = file.WriteString("timestamp,open,high,low,close,volume\n")

	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < count; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Hour)
		price := 100.0 + float64(i)
		_, _ = file.WriteString(fmt.Sprintf("%s,%.2f,%.2f,%.2f,%.2f,%.2f\n",
			ts.Format(time.RFC3339),
			price, price+2, price-1, price+1, 1000.0+float64(i*10)))
	}

	return csvPath
}

func TestNewTransformedFeed(t *testing.T) {
	csvPath := createTestCSV(t, 10)
	config := csv.DefaultCSVConfig("TEST", "1h")

	feed, err := csv.NewCSVDataFeed(csvPath, config)
	if err != nil {
		t.Fatalf("Failed to create CSV feed: %v", err)
	}
	defer feed.Close()

	chain := NewTransformChain(NewNormalizeTransform("close"))

	tfeed, err := NewTransformedFeed(feed, chain, 5)
	if err != nil {
		t.Fatalf("NewTransformedFeed failed: %v", err)
	}

	if tfeed == nil {
		t.Error("Expected non-nil transformed feed")
	}
}

func TestTransformedFeed_Next(t *testing.T) {
	csvPath := createTestCSV(t, 50)
	config := csv.DefaultCSVConfig("TEST", "1h")

	feed, err := csv.NewCSVDataFeed(csvPath, config)
	if err != nil {
		t.Fatalf("Failed to create CSV feed: %v", err)
	}
	defer feed.Close()

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	tfeed, err := NewTransformedFeed(feed, chain, 10)
	if err != nil {
		t.Fatalf("NewTransformedFeed failed: %v", err)
	}

	candle, err := tfeed.Next()
	if err != nil {
		t.Fatalf("Next() failed: %v", err)
	}

	expected := 202.0
	if !almostEqual(candle.Close, expected, 0.01) {
		t.Errorf("Transform not applied: got %.2f, want %.2f", candle.Close, expected)
	}
}

func TestTransformedFeed_ReadAll(t *testing.T) {
	csvPath := createTestCSV(t, 50)
	config := csv.DefaultCSVConfig("TEST", "1h")

	feed, err := csv.NewCSVDataFeed(csvPath, config)
	if err != nil {
		t.Fatalf("Failed to create CSV feed: %v", err)
	}
	defer feed.Close()

	chain := NewTransformChain(NewNormalizeTransform("close"))

	tfeed, err := NewTransformedFeed(feed, chain, 10)
	if err != nil {
		t.Fatalf("NewTransformedFeed failed: %v", err)
	}

	candles, err := tfeed.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() failed: %v", err)
	}

	if len(candles) != 50 {
		t.Errorf("Expected 50 candles, got %d", len(candles))
	}

	if !almostEqual(candles[0].Close, 0.0, 0.01) {
		t.Errorf("First candle should be ~0, got %.4f", candles[0].Close)
	}

	if !almostEqual(candles[len(candles)-1].Close, 1.0, 0.01) {
		t.Errorf("Last candle should be ~1, got %.4f", candles[len(candles)-1].Close)
	}
}

func TestTransformedFeed_ComplexChain(t *testing.T) {
	csvPath := createTestCSV(t, 100)
	config := csv.DefaultCSVConfig("TEST", "1h")

	feed, err := csv.NewCSVDataFeed(csvPath, config)
	if err != nil {
		t.Fatalf("Failed to create CSV feed: %v", err)
	}
	defer feed.Close()

	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewMovingAverageSmoothTransform(5, "close"),
		NewScaleTransform(100.0, "close"),
	)

	tfeed, err := NewTransformedFeed(feed, chain, 20)
	if err != nil {
		t.Fatalf("NewTransformedFeed failed: %v", err)
	}

	candles, err := tfeed.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() failed: %v", err)
	}

	if len(candles) != 100 {
		t.Errorf("Expected 100 candles, got %d", len(candles))
	}

	for i, candle := range candles {
		if candle.Close < 0 || candle.Close > 100 {
			t.Errorf("Candle %d close %.2f out of range [0, 100]", i, candle.Close)
		}
	}
}

func TestTransformedFeed_InvalidChain(t *testing.T) {
	csvPath := createTestCSV(t, 10)
	config := csv.DefaultCSVConfig("TEST", "1h")

	feed, err := csv.NewCSVDataFeed(csvPath, config)
	if err != nil {
		t.Fatalf("Failed to create CSV feed: %v", err)
	}
	defer feed.Close()

	chain := NewTransformChain(NewMovingAverageSmoothTransform(0, "close"))

	_, err = NewTransformedFeed(feed, chain, 5)
	if err == nil {
		t.Error("Expected error for invalid chain")
	}
}

func TestStatsTracker(t *testing.T) {
	csvPath := createTestCSV(t, 20)
	config := csv.DefaultCSVConfig("TEST", "1h")

	feed, err := csv.NewCSVDataFeed(csvPath, config)
	if err != nil {
		t.Fatalf("Failed to create CSV feed: %v", err)
	}
	defer feed.Close()

	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	tfeed, err := NewTransformedFeed(feed, chain, 5)
	if err != nil {
		t.Fatalf("NewTransformedFeed failed: %v", err)
	}

	tracker := NewStatsTracker(tfeed)

	for i := 0; i < 10; i++ {
		_, err := tracker.Next()
		if err != nil {
			t.Fatalf("Next() failed: %v", err)
		}
	}

	stats := tracker.Stats()
	if stats.CandlesRead != 10 {
		t.Errorf("Expected 10 candles read, got %d", stats.CandlesRead)
	}

	tracker.ResetStats()
	stats = tracker.Stats()
	if stats.CandlesRead != 0 {
		t.Errorf("Expected 0 after reset, got %d", stats.CandlesRead)
	}
}
