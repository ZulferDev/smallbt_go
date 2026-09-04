package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/1jehuang/backtest/internal/market"
)

var ErrNoMoreData = errors.New("no more data")

// CSVFeed reads OHLCV data from a CSV file.
// Expected CSV format: timestamp,open,high,low,close,volume
// First row can be headers (automatically detected and skipped).
type CSVFeed struct {
	symbol    market.Symbol
	timeframe market.Timeframe
	candles   []market.Candle
	position  int
}

// CSVConfig holds configuration for CSV parsing.
type CSVConfig struct {
	Symbol       market.Symbol
	Timeframe    market.Timeframe
	HasHeaders   bool
	TimestampCol int // 0-based index
	OpenCol      int
	HighCol      int
	LowCol       int
	CloseCol     int
	VolumeCol    int
	TimeFormat   string // If empty, tries common formats
}

// DefaultCSVConfig returns the default CSV configuration.
// Assumes format: timestamp,open,high,low,close,volume
func DefaultCSVConfig(symbol market.Symbol, timeframe market.Timeframe) CSVConfig {
	return CSVConfig{
		Symbol:       symbol,
		Timeframe:    timeframe,
		HasHeaders:   true,
		TimestampCol: 0,
		OpenCol:      1,
		HighCol:      2,
		LowCol:       3,
		CloseCol:     4,
		VolumeCol:    5,
		TimeFormat:   "", // Auto-detect
	}
}

// NewCSVFeed creates a new CSV data feed from a file.
func NewCSVFeed(filename string, config CSVConfig) (*CSVFeed, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Skip headers if present
	startIdx := 0
	if config.HasHeaders {
		startIdx = 1
	}

	candles := make([]market.Candle, 0, len(records)-startIdx)

	for i := startIdx; i < len(records); i++ {
		record := records[i]

		// Validate record length
		maxCol := max(config.TimestampCol, config.OpenCol, config.HighCol,
			config.LowCol, config.CloseCol, config.VolumeCol)
		if len(record) <= maxCol {
			return nil, fmt.Errorf("row %d: insufficient columns (expected at least %d, got %d)",
				i+1, maxCol+1, len(record))
		}

		candle, err := parseCSVRow(record, config)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+1, err)
		}

		// Validate candle
		if !candle.IsValid() {
			return nil, fmt.Errorf("row %d: invalid OHLC relationships", i+1)
		}

		candles = append(candles, candle)
	}

	// Validate chronological order
	if err := validateChronological(candles); err != nil {
		return nil, fmt.Errorf("chronological validation failed: %w", err)
	}

	return &CSVFeed{
		symbol:    config.Symbol,
		timeframe: config.Timeframe,
		candles:   candles,
		position:  0,
	}, nil
}

// parseCSVRow parses a single CSV row into a Candle.
func parseCSVRow(record []string, config CSVConfig) (market.Candle, error) {
	// Parse timestamp
	timestampStr := strings.TrimSpace(record[config.TimestampCol])
	var timestamp time.Time
	var err error

	if config.TimeFormat != "" {
		timestamp, err = time.Parse(config.TimeFormat, timestampStr)
	} else {
		timestamp, err = parseTimestamp(timestampStr)
	}
	if err != nil {
		return market.Candle{}, fmt.Errorf("parse timestamp: %w", err)
	}

	// Parse OHLCV
	open, err := strconv.ParseFloat(strings.TrimSpace(record[config.OpenCol]), 64)
	if err != nil {
		return market.Candle{}, fmt.Errorf("parse open: %w", err)
	}

	high, err := strconv.ParseFloat(strings.TrimSpace(record[config.HighCol]), 64)
	if err != nil {
		return market.Candle{}, fmt.Errorf("parse high: %w", err)
	}

	low, err := strconv.ParseFloat(strings.TrimSpace(record[config.LowCol]), 64)
	if err != nil {
		return market.Candle{}, fmt.Errorf("parse low: %w", err)
	}

	close, err := strconv.ParseFloat(strings.TrimSpace(record[config.CloseCol]), 64)
	if err != nil {
		return market.Candle{}, fmt.Errorf("parse close: %w", err)
	}

	volume, err := strconv.ParseFloat(strings.TrimSpace(record[config.VolumeCol]), 64)
	if err != nil {
		return market.Candle{}, fmt.Errorf("parse volume: %w", err)
	}

	return market.Candle{
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}, nil
}

// parseTimestamp tries common timestamp formats.
func parseTimestamp(s string) (time.Time, error) {
	// Try Unix timestamp first (numeric)
	if unix, err := strconv.ParseInt(s, 10, 64); err == nil {
		// Determine if it's seconds or milliseconds based on magnitude
		if unix > 1e12 {
			// Milliseconds
			return time.Unix(unix/1000, (unix%1000)*1e6).UTC(), nil
		}
		// Seconds
		return time.Unix(unix, 0).UTC(), nil
	}

	// Try common string formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"01/02/2006 15:04:05",
		"01/02/2006",
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t.UTC(), nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("unrecognized timestamp format %q: %w", s, lastErr)
}

// validateChronological checks that candles are in chronological order.
func validateChronological(candles []market.Candle) error {
	for i := 1; i < len(candles); i++ {
		if !candles[i].Timestamp.After(candles[i-1].Timestamp) {
			return fmt.Errorf("candle at index %d is not after previous candle (times: %v, %v)",
				i, candles[i-1].Timestamp, candles[i].Timestamp)
		}
	}
	return nil
}

// Next returns the next candle as MarketData.
func (f *CSVFeed) Next() (market.MarketData, error) {
	if f.position >= len(f.candles) {
		return market.MarketData{}, ErrNoMoreData
	}

	candle := f.candles[f.position]
	f.position++

	// Return MarketData with single candle
	md := market.MarketData{
		Symbol:    f.symbol,
		Timeframe: f.timeframe,
		Candles:   []market.Candle{candle},
	}

	return md, nil
}

// Reset resets the feed to the beginning.
func (f *CSVFeed) Reset() {
	f.position = 0
}

// Symbol returns the symbol.
func (f *CSVFeed) Symbol() market.Symbol {
	return f.symbol
}

// Timeframe returns the timeframe.
func (f *CSVFeed) Timeframe() market.Timeframe {
	return f.timeframe
}

// Length returns the total number of candles.
func (f *CSVFeed) Length() int {
	return len(f.candles)
}

// Position returns the current position.
func (f *CSVFeed) Position() int {
	return f.position
}

// GetCandles returns all candles (for testing/debugging).
func (f *CSVFeed) GetCandles() []market.Candle {
	return f.candles
}
