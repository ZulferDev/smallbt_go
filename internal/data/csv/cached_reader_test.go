package csv

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/data/cache"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestNewCachedCSVFeed(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
	}
	if err := writeTestCSV(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	feed, err := NewCachedCSVFeed(testFile, config, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedCSVFeed failed: %v", err)
	}

	if feed.feed == nil {
		t.Error("Expected feed to be initialized")
	}
	if feed.cache == nil {
		t.Error("Expected cache to be initialized")
	}
	if feed.symbol != "BTC" {
		t.Errorf("Expected symbol BTC, got %s", feed.symbol)
	}
	if feed.timeframe != "1h" {
		t.Errorf("Expected timeframe 1h, got %s", feed.timeframe)
	}
}

func TestNewCachedCSVFeed_NilCache(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
	}
	if err := writeTestCSV(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	feed, err := NewCachedCSVFeed(testFile, config, nil, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedCSVFeed failed: %v", err)
	}

	if feed.cache == nil {
		t.Error("Expected no-op cache to be initialized")
	}
}

func TestNewCachedCSVFeed_InvalidFile(t *testing.T) {
	c := cache.NewLRUCache(10)
	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	_, err := NewCachedCSVFeed("nonexistent.csv", config, c, "BTC", "1h")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCachedCSVFeed_ReadAll_CacheMiss(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test file with 3 candles
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), Open: 102, High: 106, Low: 100, Close: 104, Volume: 1100},
		{Timestamp: time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC), Open: 104, High: 108, Low: 102, Close: 106, Volume: 1200},
	}
	if err := writeTestCSV(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	feed, err := NewCachedCSVFeed(testFile, config, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedCSVFeed failed: %v", err)
	}

	// First read - cache miss
	result, err := feed.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 candles, got %d", len(result))
	}

	// Verify cache stats - ReadRange internally calls ReadAll
	stats := feed.Stats()
	if stats.Misses < 1 {
		t.Errorf("Expected at least 1 cache miss, got %d", stats.Misses)
	}
	if stats.Hits != 0 {
		t.Errorf("Expected 0 cache hits, got %d", stats.Hits)
	}
}

func TestCachedCSVFeed_ReadAll_CacheHit(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), Open: 102, High: 106, Low: 100, Close: 104, Volume: 1100},
	}
	if err := writeTestCSV(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	feed, err := NewCachedCSVFeed(testFile, config, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedCSVFeed failed: %v", err)
	}

	// First read - cache miss
	result1, err := feed.ReadAll()
	if err != nil {
		t.Fatalf("First ReadAll failed: %v", err)
	}

	// Second read - cache hit
	result2, err := feed.ReadAll()
	if err != nil {
		t.Fatalf("Second ReadAll failed: %v", err)
	}

	// Results should be identical
	if len(result1) != len(result2) {
		t.Errorf("Expected same length, got %d and %d", len(result1), len(result2))
	}

	// Verify cache stats
	stats := feed.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
}

func TestCachedCSVFeed_ReadRange_CacheMiss(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test file with 5 candles
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), Open: 102, High: 106, Low: 100, Close: 104, Volume: 1100},
		{Timestamp: time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC), Open: 104, High: 108, Low: 102, Close: 106, Volume: 1200},
		{Timestamp: time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC), Open: 106, High: 110, Low: 104, Close: 108, Volume: 1300},
		{Timestamp: time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC), Open: 108, High: 112, Low: 106, Close: 110, Volume: 1400},
	}
	if err := writeTestCSV(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	feed, err := NewCachedCSVFeed(testFile, config, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedCSVFeed failed: %v", err)
	}

	// Read range - cache miss
	start := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC)
	result, err := feed.ReadRange(start, end)
	if err != nil {
		t.Fatalf("ReadRange failed: %v", err)
	}

	// Should return 3 candles (1h, 2h, 3h)
	if len(result) != 3 {
		t.Errorf("Expected 3 candles, got %d", len(result))
	}

	// Verify cache stats - ReadRange internally calls ReadAll
	stats := feed.Stats()
	if stats.Misses < 1 {
		t.Errorf("Expected at least 1 cache miss, got %d", stats.Misses)
	}
}

func TestCachedCSVFeed_ReadRange_CacheHit(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), Open: 102, High: 106, Low: 100, Close: 104, Volume: 1100},
		{Timestamp: time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC), Open: 104, High: 108, Low: 102, Close: 106, Volume: 1200},
	}
	if err := writeTestCSV(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	feed, err := NewCachedCSVFeed(testFile, config, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedCSVFeed failed: %v", err)
	}

	// First range read - cache miss
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	_, err = feed.ReadRange(start, end)
	if err != nil {
		t.Fatalf("First ReadRange failed: %v", err)
	}

	// Second range read (different range) - cache hit
	start2 := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	end2 := time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC)
	result2, err := feed.ReadRange(start2, end2)
	if err != nil {
		t.Fatalf("Second ReadRange failed: %v", err)
	}

	if len(result2) != 2 {
		t.Errorf("Expected 2 candles, got %d", len(result2))
	}

	// Verify cache stats - first ReadRange creates cache, second reuses it
	stats := feed.Stats()
	if stats.Hits < 1 {
		t.Errorf("Expected at least 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses < 1 {
		t.Errorf("Expected at least 1 cache miss, got %d", stats.Misses)
	}
}

func TestCachedCSVFeed_ClearCache(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
	}
	if err := writeTestCSV(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	feed, err := NewCachedCSVFeed(testFile, config, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedCSVFeed failed: %v", err)
	}

	// Read to populate cache
	_, err = feed.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	// Clear cache
	feed.ClearCache()

	// Next read should be cache miss
	_, err = feed.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll after clear failed: %v", err)
	}

	stats := feed.Stats()
	if stats.Misses != 2 {
		t.Errorf("Expected 2 cache misses after clear, got %d", stats.Misses)
	}
}

func TestCachedCSVFeed_Stats(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
	}
	if err := writeTestCSV(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	config := CSVConfig{
		Symbol:       market.Symbol("BTC"),
		Timeframe:    market.Timeframe("1h"),
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
	}
	feed, err := NewCachedCSVFeed(testFile, config, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedCSVFeed failed: %v", err)
	}

	// Read multiple times
	feed.ReadAll()
	feed.ReadAll()
	feed.ReadAll()

	stats := feed.Stats()
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

// writeTestCSV writes test candles to a CSV file
func writeTestCSV(path string, candles []*market.Candle) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	if _, err := file.WriteString("timestamp,open,high,low,close,volume\n"); err != nil {
		return err
	}

	// Write candles
	for _, candle := range candles {
		line := candle.Timestamp.Format(time.RFC3339) + "," +
			formatFloat(candle.Open) + "," +
			formatFloat(candle.High) + "," +
			formatFloat(candle.Low) + "," +
			formatFloat(candle.Close) + "," +
			formatFloat(candle.Volume) + "\n"
		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.8f", f)
}
