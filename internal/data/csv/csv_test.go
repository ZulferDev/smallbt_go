package csv

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/market"
)

func TestNewCSVFeed(t *testing.T) {
	// Create temporary CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test.csv")

	// Write test data
	content := `timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,100.0,105.0,99.0,102.0,1000.0
2024-01-01T01:00:00Z,102.0,108.0,101.0,107.0,1200.0
2024-01-01T02:00:00Z,107.0,110.0,106.0,109.0,800.0
`
	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create feed
	config := DefaultCSVConfig("BTCUSDT", market.Timeframe1h)
	feed, err := NewCSVFeed(csvFile, config)
	if err != nil {
		t.Fatalf("NewCSVFeed() error = %v", err)
	}

	if feed.Symbol() != "BTCUSDT" {
		t.Errorf("expected symbol BTCUSDT, got %v", feed.Symbol())
	}

	if feed.Timeframe() != market.Timeframe1h {
		t.Errorf("expected timeframe 1h, got %v", feed.Timeframe())
	}

	if feed.Length() != 3 {
		t.Errorf("expected 3 candles, got %d", feed.Length())
	}
}

func TestCSVFeedDeterministicIteration(t *testing.T) {
	// Create temporary CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test.csv")

	content := `timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,100.0,105.0,99.0,102.0,1000.0
2024-01-01T01:00:00Z,102.0,108.0,101.0,107.0,1200.0
2024-01-01T02:00:00Z,107.0,110.0,106.0,109.0,800.0
`
	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	config := DefaultCSVConfig("BTCUSDT", market.Timeframe1h)
	feed, err := NewCSVFeed(csvFile, config)
	if err != nil {
		t.Fatalf("NewCSVFeed() error = %v", err)
	}

	// First iteration - collect all candles
	var firstRun []market.Candle
	for {
		md, err := feed.Next()
		if err == ErrNoMoreData {
			break
		}
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		firstRun = append(firstRun, md.Candles[0])
	}

	// Reset and iterate again
	feed.Reset()

	var secondRun []market.Candle
	for {
		md, err := feed.Next()
		if err == ErrNoMoreData {
			break
		}
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		secondRun = append(secondRun, md.Candles[0])
	}

	// Verify deterministic behavior
	if len(firstRun) != len(secondRun) {
		t.Fatalf("iteration count mismatch: first=%d, second=%d", len(firstRun), len(secondRun))
	}

	for i := range firstRun {
		if !firstRun[i].Timestamp.Equal(secondRun[i].Timestamp) {
			t.Errorf("candle %d: timestamp mismatch", i)
		}
		if firstRun[i].Open != secondRun[i].Open {
			t.Errorf("candle %d: open mismatch", i)
		}
	}
}

func TestCSVFeedChronologicalValidation(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test.csv")

	// Out of order data
	content := `timestamp,open,high,low,close,volume
2024-01-01T02:00:00Z,107.0,110.0,106.0,109.0,800.0
2024-01-01T00:00:00Z,100.0,105.0,99.0,102.0,1000.0
2024-01-01T01:00:00Z,102.0,108.0,101.0,107.0,1200.0
`
	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	config := DefaultCSVConfig("BTCUSDT", market.Timeframe1h)
	_, err := NewCSVFeed(csvFile, config)
	if err == nil {
		t.Error("expected error for non-chronological data")
	}
}

func TestCSVFeedOHLCValidation(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test.csv")

	// Invalid OHLC: high < low
	content := `timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,100.0,99.0,105.0,102.0,1000.0
`
	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	config := DefaultCSVConfig("BTCUSDT", market.Timeframe1h)
	_, err := NewCSVFeed(csvFile, config)
	if err == nil {
		t.Error("expected error for invalid OHLC")
	}
}

func TestCSVFeedUnixTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test.csv")

	// Unix timestamp in seconds
	content := `timestamp,open,high,low,close,volume
1704067200,100.0,105.0,99.0,102.0,1000.0
1704070800,102.0,108.0,101.0,107.0,1200.0
`
	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	config := DefaultCSVConfig("BTCUSDT", market.Timeframe1h)
	feed, err := NewCSVFeed(csvFile, config)
	if err != nil {
		t.Fatalf("NewCSVFeed() error = %v", err)
	}

	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	md, err := feed.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	if !md.Candles[0].Timestamp.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, md.Candles[0].Timestamp)
	}
}

func TestCSVFeedPosition(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "test.csv")

	content := `timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,100.0,105.0,99.0,102.0,1000.0
2024-01-01T01:00:00Z,102.0,108.0,101.0,107.0,1200.0
2024-01-01T02:00:00Z,107.0,110.0,106.0,109.0,800.0
`
	if err := os.WriteFile(csvFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	config := DefaultCSVConfig("BTCUSDT", market.Timeframe1h)
	feed, err := NewCSVFeed(csvFile, config)
	if err != nil {
		t.Fatalf("NewCSVFeed() error = %v", err)
	}

	if feed.Position() != 0 {
		t.Errorf("expected initial position 0, got %d", feed.Position())
	}

	_, _ = feed.Next()
	if feed.Position() != 1 {
		t.Errorf("expected position 1 after first Next(), got %d", feed.Position())
	}

	_, _ = feed.Next()
	if feed.Position() != 2 {
		t.Errorf("expected position 2 after second Next(), got %d", feed.Position())
	}

	feed.Reset()
	if feed.Position() != 0 {
		t.Errorf("expected position 0 after Reset(), got %d", feed.Position())
	}
}
