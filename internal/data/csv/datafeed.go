package csv

import (
	"context"
	"io"

	"github.com/ZulferDev/smallbt_go/internal/data"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CSVDataFeed implements data.DataFeed interface for CSV files
// It wraps the existing CSVFeed to provide the new interface
type CSVDataFeed struct {
	feed   *CSVFeed
	closed bool
}

// NewCSVDataFeed creates a DataFeed from a CSV file
func NewCSVDataFeed(filename string, config CSVConfig) (*CSVDataFeed, error) {
	feed, err := NewCSVFeed(filename, config)
	if err != nil {
		return nil, err
	}

	return &CSVDataFeed{
		feed:   feed,
		closed: false,
	}, nil
}

// Subscribe implements data.DataFeed interface
// For CSV, this is a no-op (file already loaded in constructor)
func (f *CSVDataFeed) Subscribe(ctx context.Context, symbols []string) error {
	if f.closed {
		return data.ErrFeedClosed
	}

	// CSV feed is single-symbol, already loaded
	// Validate that requested symbol matches
	if len(symbols) > 0 && symbols[0] != string(f.feed.Symbol()) {
		return data.ErrSymbolNotSupported
	}

	return nil
}

// Next implements data.DataFeed interface
// Returns the next candle chronologically, or io.EOF when done
func (f *CSVDataFeed) Next(ctx context.Context) (*market.Candle, error) {
	if f.closed {
		return nil, data.ErrFeedClosed
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get next candle from underlying feed
	md, err := f.feed.Next()
	if err != nil {
		if err == ErrNoMoreData {
			return nil, io.EOF
		}
		return nil, err
	}

	// Extract candle from MarketData
	if len(md.Candles) == 0 {
		return nil, data.ErrNoData
	}

	return &md.Candles[0], nil
}

// Close implements data.DataFeed interface
func (f *CSVDataFeed) Close() error {
	f.closed = true
	return nil
}

// Reset resets the feed to the beginning (useful for testing)
func (f *CSVDataFeed) Reset() {
	f.feed.Reset()
}

// Symbol returns the symbol this feed provides
func (f *CSVDataFeed) Symbol() market.Symbol {
	return f.feed.Symbol()
}

// Timeframe returns the timeframe
func (f *CSVDataFeed) Timeframe() market.Timeframe {
	return f.feed.Timeframe()
}

// Length returns total number of candles
func (f *CSVDataFeed) Length() int {
	return f.feed.Length()
}

// Verify interface compliance at compile time
var _ data.DataFeed = (*CSVDataFeed)(nil)
