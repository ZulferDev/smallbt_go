package parquet

import (
	"time"

	"github.com/ZulferDev/smallbt_go/internal/data/cache"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CachedParquetReader wraps ParquetReader with caching layer.
// Caches validated candle data to avoid redundant file reads and validation.
type CachedParquetReader struct {
	reader    *ParquetReader
	cache     cache.Cache
	symbol    string
	timeframe string
}

// NewCachedParquetReader creates a new CachedParquetReader.
// If cache is nil, creates a no-op cache (no caching).
func NewCachedParquetReader(path string, c cache.Cache, symbol, timeframe string) (*CachedParquetReader, error) {
	reader, err := NewParquetReader(path)
	if err != nil {
		return nil, err
	}

	if c == nil {
		c = cache.NewNoOpCache()
	}

	return &CachedParquetReader{
		reader:    reader,
		cache:     c,
		symbol:    symbol,
		timeframe: timeframe,
	}, nil
}

// Read reads all candles from the Parquet file.
// Uses cache if available, otherwise reads from file and caches result.
func (r *CachedParquetReader) Read() ([]*market.Candle, error) {
	// Use file path as part of cache key for consistency
	cacheKey := r.symbol + ":" + r.timeframe + ":full"

	// Try cache first
	cached := r.cache.Get(cacheKey)
	if cached != nil {
		return cached, nil
	}

	// Cache miss - read from file
	candles, err := r.reader.Read()
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Put(cacheKey, candles)

	return candles, nil
}

// ReadRange reads candles within a specific time range.
// Uses cache if available, otherwise reads from file and caches result.
func (r *CachedParquetReader) ReadRange(start, end time.Time) ([]*market.Candle, error) {
	// For range queries, we cache the full dataset then filter
	cacheKey := r.symbol + ":" + r.timeframe + ":full"

	// Try cache first
	cached := r.cache.Get(cacheKey)
	var allCandles []*market.Candle
	var err error

	if cached != nil {
		allCandles = cached
	} else {
		// Cache miss - read from file
		allCandles, err = r.reader.Read()
		if err != nil {
			return nil, err
		}

		// Cache the full result
		r.cache.Put(cacheKey, allCandles)
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

// Close closes the underlying Parquet reader and releases resources.
func (r *CachedParquetReader) Close() error {
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
}

// Stats returns cache statistics.
func (r *CachedParquetReader) Stats() cache.CacheStats {
	return r.cache.Stats()
}

// ClearCache clears the cache.
func (r *CachedParquetReader) ClearCache() {
	r.cache.Clear()
}
