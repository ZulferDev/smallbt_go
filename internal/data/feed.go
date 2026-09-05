package data

import (
	"fmt"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// ErrNoMoreData is returned when the data feed has no more data to provide.
var ErrNoMoreData = fmt.Errorf("no more data")

// DataFeed represents a source of market data.
type DataFeed interface {
	// Next returns the next market event from the feed.
	// Returns ErrNoMoreData when the feed is exhausted.
	Next() (market.MarketData, error)

	// Reset resets the feed to the beginning.
	Reset()

	// Symbol returns the symbol being traded.
	Symbol() market.Symbol

	// Timeframe returns the timeframe of the data.
	Timeframe() market.Timeframe

	// Length returns the total number of candles available.
	Length() int

	// Position returns the current position in the feed.
	Position() int
}
