package data

import (
	"context"
	"errors"
	"io"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// DataFeed provides market data (historical or real-time)
// Implementations include CSV, Parquet, WebSocket, REST API
type DataFeed interface {
	// Subscribe prepares the feed for the given symbols
	// For CSV/Parquet: validates files exist
	// For WebSocket: establishes connection and subscribes
	// For REST: validates API credentials
	Subscribe(ctx context.Context, symbols []string) error

	// Next returns the next candle chronologically
	// Returns io.EOF when no more data available
	// For backtest: returns next historical candle
	// For real-time: blocks until next candle arrives
	Next(ctx context.Context) (*market.Candle, error)

	// Close releases resources (file handles, connections, etc)
	Close() error
}

// FeedType distinguishes data source types
type FeedType string

const (
	// FeedTypeCSV reads historical data from CSV files
	FeedTypeCSV FeedType = "csv"

	// FeedTypeParquet reads historical data from Parquet files
	FeedTypeParquet FeedType = "parquet"

	// FeedTypeWebSocket receives real-time data via WebSocket
	FeedTypeWebSocket FeedType = "websocket"

	// FeedTypeREST polls real-time data via REST API
	FeedTypeREST FeedType = "rest"
)

var (
	// ErrNoData is returned when feed has no data
	ErrNoData = errors.New("no data available")

	// ErrInvalidFormat is returned when data format is invalid
	ErrInvalidFormat = errors.New("invalid data format")

	// ErrConnectionLost is returned when WebSocket/network connection drops
	ErrConnectionLost = errors.New("connection lost")

	// ErrFeedClosed is returned when operation attempted on closed feed
	ErrFeedClosed = errors.New("feed is closed")

	// ErrSymbolNotSupported is returned when symbol is not available
	ErrSymbolNotSupported = errors.New("symbol not supported")

	// ErrRateLimited is returned when API rate limit exceeded
	ErrRateLimited = errors.New("rate limit exceeded")
)

// IsRetryable returns true if the error indicates a transient failure
// that should be retried (connection issues, rate limits, etc)
func IsRetryable(err error) bool {
	return errors.Is(err, ErrConnectionLost) || errors.Is(err, ErrRateLimited)
}

// IsEndOfData returns true if error indicates normal end of data stream
func IsEndOfData(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, ErrNoData)
}
