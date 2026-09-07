package transform

import (
	"fmt"
	"sync"

	"github.com/ZulferDev/smallbt_go/internal/data"
	"github.com/ZulferDev/smallbt_go/internal/data/cache"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// CachedTransformedFeed adds caching layer to transformed feed.
type CachedTransformedFeed struct {
	feed  *TransformedFeed
	cache cache.Cache
	key   string
	mu    sync.RWMutex
}

// NewCachedTransformedFeed creates a cached transformed feed.
func NewCachedTransformedFeed(feed data.DataFeed, chain *TransformChain, cache cache.Cache, cacheKey string, batchSize int) (*CachedTransformedFeed, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache cannot be nil")
	}
	if cacheKey == "" {
		return nil, fmt.Errorf("cache key cannot be empty")
	}

	tfeed, err := NewTransformedFeed(feed, chain, batchSize)
	if err != nil {
		return nil, err
	}

	return &CachedTransformedFeed{
		feed:  tfeed,
		cache: cache,
		key:   cacheKey,
	}, nil
}

// ReadAll reads all data with caching.
func (ctf *CachedTransformedFeed) ReadAll() ([]*market.Candle, error) {
	// Try cache first
	ctf.mu.RLock()
	if cached := ctf.cache.Get(ctf.key); cached != nil {
		ctf.mu.RUnlock()
		return cached, nil
	}
	ctf.mu.RUnlock()

	// Cache miss - read and transform
	candles, err := ctf.feed.ReadAll()
	if err != nil {
		return nil, err
	}

	// Store in cache
	ctf.mu.Lock()
	ctf.cache.Put(ctf.key, candles)
	ctf.mu.Unlock()

	return candles, nil
}

// Next reads next candle (no caching for streaming).
func (ctf *CachedTransformedFeed) Next() (*market.Candle, error) {
	return ctf.feed.Next()
}

// InvalidateCache clears cached data.
func (ctf *CachedTransformedFeed) InvalidateCache() {
	ctf.mu.Lock()
	defer ctf.mu.Unlock()
	ctf.cache.Clear()
}

// Close closes the underlying feed.
func (ctf *CachedTransformedFeed) Close() error {
	return ctf.feed.Close()
}

// CacheStats returns cache statistics.
func (ctf *CachedTransformedFeed) CacheStats() cache.CacheStats {
	return ctf.cache.Stats()
}

// TransformCache is a specialized cache for transformed data.
type TransformCache struct {
	cache cache.Cache
	mu    sync.RWMutex
}

// NewTransformCache creates a transform cache with given capacity.
func NewTransformCache(capacity int) *TransformCache {
	return &TransformCache{
		cache: cache.NewLRUCache(capacity),
	}
}

// GenerateKey generates a cache key from symbol, timeframe, and transform chain.
func (tc *TransformCache) GenerateKey(symbol, timeframe string, chain *TransformChain) string {
	// Simple key generation: symbol:timeframe:chainName
	return fmt.Sprintf("%s:%s:%s", symbol, timeframe, chain.Name())
}

// Get retrieves cached transformed data.
func (tc *TransformCache) Get(key string) ([]*market.Candle, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if value := tc.cache.Get(key); value != nil {
		return value, true
	}
	return nil, false
}

// Put stores transformed data in cache.
func (tc *TransformCache) Put(key string, candles []*market.Candle) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.cache.Put(key, candles)
}

// Clear clears all cached data.
func (tc *TransformCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.cache.Clear()
}

// Stats returns cache statistics.
func (tc *TransformCache) Stats() cache.CacheStats {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.cache.Stats()
}

// PreloadCache preloads transformed data into cache.
type PreloadCache struct {
	cache *TransformCache
	feeds map[string]*TransformedFeed
	mu    sync.RWMutex
}

// NewPreloadCache creates a preload cache.
func NewPreloadCache(capacity int) *PreloadCache {
	return &PreloadCache{
		cache: NewTransformCache(capacity),
		feeds: make(map[string]*TransformedFeed),
	}
}

// Preload loads and caches data for given symbol/feed.
func (pc *PreloadCache) Preload(symbol, timeframe string, feed data.DataFeed, chain *TransformChain, batchSize int) error {
	tfeed, err := NewTransformedFeed(feed, chain, batchSize)
	if err != nil {
		return fmt.Errorf("create transformed feed: %w", err)
	}

	// Read all data
	candles, err := tfeed.ReadAll()
	if err != nil {
		tfeed.Close()
		return fmt.Errorf("read all: %w", err)
	}

	// Generate key and cache
	key := pc.cache.GenerateKey(symbol, timeframe, chain)
	pc.cache.Put(key, candles)

	// Store feed reference
	pc.mu.Lock()
	pc.feeds[key] = tfeed
	pc.mu.Unlock()

	return nil
}

// Get retrieves preloaded data.
func (pc *PreloadCache) Get(symbol, timeframe string, chain *TransformChain) ([]*market.Candle, bool) {
	key := pc.cache.GenerateKey(symbol, timeframe, chain)
	return pc.cache.Get(key)
}

// Clear clears preloaded cache.
func (pc *PreloadCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Close all feeds
	for _, tfeed := range pc.feeds {
		tfeed.Close()
	}
	pc.feeds = make(map[string]*TransformedFeed)
	pc.cache.Clear()
}

// Stats returns cache statistics.
func (pc *PreloadCache) Stats() cache.CacheStats {
	return pc.cache.Stats()
}

// SmartCache automatically caches frequently accessed transforms.
type SmartCache struct {
	cache       *TransformCache
	accessCount map[string]int
	threshold   int
	mu          sync.RWMutex
}

// NewSmartCache creates a smart cache with access threshold.
func NewSmartCache(capacity, accessThreshold int) *SmartCache {
	return &SmartCache{
		cache:       NewTransformCache(capacity),
		accessCount: make(map[string]int),
		threshold:   accessThreshold,
	}
}

// GetOrCompute gets from cache or computes and caches if accessed frequently.
func (sc *SmartCache) GetOrCompute(key string, compute func() ([]*market.Candle, error)) ([]*market.Candle, error) {
	// Try cache first
	if candles, ok := sc.cache.Get(key); ok {
		return candles, nil
	}

	// Track access
	sc.mu.Lock()
	sc.accessCount[key]++
	count := sc.accessCount[key]
	sc.mu.Unlock()

	// Compute
	candles, err := compute()
	if err != nil {
		return nil, err
	}

	// Cache if accessed frequently
	if count >= sc.threshold {
		sc.cache.Put(key, candles)
	}

	return candles, nil
}

// ResetAccessCount resets access tracking.
func (sc *SmartCache) ResetAccessCount() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.accessCount = make(map[string]int)
}

// Stats returns cache statistics.
func (sc *SmartCache) Stats() cache.CacheStats {
	return sc.cache.Stats()
}

// AccessStats returns access count statistics.
func (sc *SmartCache) AccessStats() map[string]int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	stats := make(map[string]int, len(sc.accessCount))
	for k, v := range sc.accessCount {
		stats[k] = v
	}
	return stats
}
