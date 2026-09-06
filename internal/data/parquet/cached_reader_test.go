package parquet

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/data/cache"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestNewCachedParquetReader(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
	}
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	reader, err := NewCachedParquetReader(testFile, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedParquetReader failed: %v", err)
	}
	defer reader.Close()

	if reader.reader == nil {
		t.Error("Expected reader to be initialized")
	}
	if reader.cache == nil {
		t.Error("Expected cache to be initialized")
	}
	if reader.symbol != "BTC" {
		t.Errorf("Expected symbol BTC, got %s", reader.symbol)
	}
	if reader.timeframe != "1h" {
		t.Errorf("Expected timeframe 1h, got %s", reader.timeframe)
	}
}

func TestNewCachedParquetReader_NilCache(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
	}
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	reader, err := NewCachedParquetReader(testFile, nil, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedParquetReader failed: %v", err)
	}
	defer reader.Close()

	if reader.cache == nil {
		t.Error("Expected no-op cache to be initialized")
	}
}

func TestNewCachedParquetReader_InvalidFile(t *testing.T) {
	c := cache.NewLRUCache(10)
	_, err := NewCachedParquetReader("nonexistent.parquet", c, "BTC", "1h")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCachedParquetReader_Read_CacheMiss(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file with 3 candles
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), Open: 102, High: 106, Low: 100, Close: 104, Volume: 1100},
		{Timestamp: time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC), Open: 104, High: 108, Low: 102, Close: 106, Volume: 1200},
	}
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	reader, err := NewCachedParquetReader(testFile, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// First read - cache miss
	result, err := reader.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 candles, got %d", len(result))
	}

	// Verify cache stats
	stats := reader.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
	if stats.Hits != 0 {
		t.Errorf("Expected 0 cache hits, got %d", stats.Hits)
	}
}

func TestCachedParquetReader_Read_CacheHit(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), Open: 102, High: 106, Low: 100, Close: 104, Volume: 1100},
	}
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	reader, err := NewCachedParquetReader(testFile, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// First read - cache miss
	result1, err := reader.Read()
	if err != nil {
		t.Fatalf("First read failed: %v", err)
	}

	// Second read - cache hit
	result2, err := reader.Read()
	if err != nil {
		t.Fatalf("Second read failed: %v", err)
	}

	// Results should be identical
	if len(result1) != len(result2) {
		t.Errorf("Expected same length, got %d and %d", len(result1), len(result2))
	}

	// Verify cache stats
	stats := reader.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
	if stats.HitRate < 0.49 || stats.HitRate > 0.51 {
		t.Errorf("Expected hit rate ~0.5, got %.2f", stats.HitRate)
	}
}

func TestCachedParquetReader_ReadRange_CacheMiss(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file with 5 candles
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), Open: 102, High: 106, Low: 100, Close: 104, Volume: 1100},
		{Timestamp: time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC), Open: 104, High: 108, Low: 102, Close: 106, Volume: 1200},
		{Timestamp: time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC), Open: 106, High: 110, Low: 104, Close: 108, Volume: 1300},
		{Timestamp: time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC), Open: 108, High: 112, Low: 106, Close: 110, Volume: 1400},
	}
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	reader, err := NewCachedParquetReader(testFile, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Read range - cache miss
	start := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC)
	result, err := reader.ReadRange(start, end)
	if err != nil {
		t.Fatalf("ReadRange failed: %v", err)
	}

	// Should return 3 candles (1h, 2h, 3h)
	if len(result) != 3 {
		t.Errorf("Expected 3 candles, got %d", len(result))
	}

	// Verify cache stats
	stats := reader.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
}

func TestCachedParquetReader_ReadRange_CacheHit(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
		{Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), Open: 102, High: 106, Low: 100, Close: 104, Volume: 1100},
		{Timestamp: time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC), Open: 104, High: 108, Low: 102, Close: 106, Volume: 1200},
	}
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	reader, err := NewCachedParquetReader(testFile, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// First range read - cache miss
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	_, err = reader.ReadRange(start, end)
	if err != nil {
		t.Fatalf("First ReadRange failed: %v", err)
	}

	// Second range read (different range) - cache hit (same full dataset)
	start2 := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	end2 := time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC)
	result2, err := reader.ReadRange(start2, end2)
	if err != nil {
		t.Fatalf("Second ReadRange failed: %v", err)
	}

	if len(result2) != 2 {
		t.Errorf("Expected 2 candles, got %d", len(result2))
	}

	// Verify cache stats - one miss, one hit
	stats := reader.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
}

func TestCachedParquetReader_ClearCache(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
	}
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	reader, err := NewCachedParquetReader(testFile, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Read to populate cache
	_, err = reader.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Clear cache
	reader.ClearCache()

	// Next read should be cache miss
	_, err = reader.Read()
	if err != nil {
		t.Fatalf("Read after clear failed: %v", err)
	}

	stats := reader.Stats()
	if stats.Misses != 2 {
		t.Errorf("Expected 2 cache misses after clear, got %d", stats.Misses)
	}
	if stats.Size != 1 {
		t.Errorf("Expected cache size 1 after re-read, got %d", stats.Size)
	}
}

func TestCachedParquetReader_Stats(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.parquet")

	// Create test file
	candles := []*market.Candle{
		{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 105, Low: 99, Close: 102, Volume: 1000},
	}
	if err := writeTestParquet(testFile, candles); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := cache.NewLRUCache(10)
	reader, err := NewCachedParquetReader(testFile, c, "BTC", "1h")
	if err != nil {
		t.Fatalf("NewCachedParquetReader failed: %v", err)
	}
	defer reader.Close()

	// Read multiple times
	reader.Read()
	reader.Read()
	reader.Read()

	stats := reader.Stats()
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.Size != 1 {
		t.Errorf("Expected cache size 1, got %d", stats.Size)
	}
}

// writeTestParquet writes test candles to a Parquet file
func writeTestParquet(path string, candles []*market.Candle) error {
	writer, err := NewParquetWriter(path)
	if err != nil {
		return err
	}
	defer writer.Close()

	if err := writer.Write(candles); err != nil {
		return err
	}

	return nil
}
