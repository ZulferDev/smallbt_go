package csv

import (
	"time"

	"github.com/ZulferDev/smallbt_go/internal/data/cache"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CachedCSVFeed wraps CSVFeed with caching layer.
// Caches validated candle data to avoid redundant file reads and validation.
type CachedCSVFeed struct {
	feed      *CSVFeed
	cache     cache.Cache
	symbol    string
	timeframe string
}

// NewCachedCSVFeed creates a new CachedCSVFeed.
// If cache is nil, creates a no-op cache (no caching).
func NewCachedCSVFeed(filename string, config CSVConfig, c cache.Cache, symbol, timeframe string) (*CachedCSVFeed, error) {
	feed, err := NewCSVFeed(filename, config)
	if err != nil {
		return nil, err
	}

	if c == nil {
		c = cache.NewNoOpCache()
	}

	return &CachedCSVFeed{
		feed:      feed,
		cache:     c,
		symbol:    symbol,
		timeframe: timeframe,
	}, nil
}

// ReadAll reads all candles from the CSV file.
// Uses cache if available, otherwise reads from file and caches result.
func (f *CachedCSVFeed) ReadAll() ([]*market.Candle, error) {
	// Use consistent cache key
	cacheKey := f.symbol + ":" + f.timeframe + ":full"

	// Try cache first
	cached := f.cache.Get(cacheKey)
	if cached != nil {
		return cached, nil
	}

	// Cache miss - read from file
	var candles []*market.Candle
	for {
		data, err := f.feed.Next()
		if err != nil {
			break
		}
		// MarketData.Candles contains the candle(s)
		for i := range data.Candles {
			candles = append(candles, &data.Candles[i])
		}
	}

	// Cache the result
	f.cache.Put(cacheKey, candles)

	return candles, nil
}

// ReadRange reads candles within a specific time range.
// Uses cache if available, otherwise reads from file and caches result.
func (f *CachedCSVFeed) ReadRange(start, end time.Time) ([]*market.Candle, error) {
	// For range queries, we cache the full dataset then filter
	cacheKey := f.symbol + ":" + f.timeframe + ":full"

	// Try cache first
	cached := f.cache.Get(cacheKey)
	var allCandles []*market.Candle
	var err error

	if cached != nil {
		allCandles = cached
	} else {
		// Cache miss - read from file
		allCandles, err = f.ReadAll()
		if err != nil {
			return nil, err
		}
	}

	// Filter by time range
	if len(allCandles) == 0 {
		return []*market.Candle{}, nil
	}

	startMillis := start.UnixMilli()
	endMillis := end.UnixMilli()

	filtered := make([]*market.Candle, 0, len(allCandles))
	for _, candle := range allCandles {
		candleMillis := candle.Timestamp.UnixMilli()
		if candleMillis >= startMillis && candleMillis <= endMillis {
			filtered = append(filtered, candle)
		}
	}

	return filtered, nil
}

// Stats returns cache statistics.
func (f *CachedCSVFeed) Stats() cache.CacheStats {
	return f.cache.Stats()
}

// ClearCache clears the cache.
func (f *CachedCSVFeed) ClearCache() {
	f.cache.Clear()
}
