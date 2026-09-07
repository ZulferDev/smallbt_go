package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Cache provides caching for validated candle data.
type Cache interface {
	// Get retrieves cached candles by key. Returns nil if not found.
	Get(key string) []*market.Candle

	// Put stores candles in cache with the given key.
	Put(key string, candles []*market.Candle)

	// Clear removes all entries from cache.
	Clear()

	// Stats returns cache statistics.
	Stats() CacheStats
}

// CacheStats holds cache statistics.
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int
	MaxSize   int
	HitRate   float64
}

// LRUCache implements Cache with Least Recently Used eviction policy.
type LRUCache struct {
	maxSize int
	mu      sync.RWMutex

	// entries maps key to cache entry
	entries map[string]*cacheEntry

	// lru is a doubly-linked list for LRU tracking
	head *cacheEntry
	tail *cacheEntry

	// stats
	hits      int64
	misses    int64
	evictions int64
}

// cacheEntry represents a cached item in the LRU list.
type cacheEntry struct {
	key      string
	candles  []*market.Candle
	prev     *cacheEntry
	next     *cacheEntry
	cachedAt time.Time
}

// NewLRUCache creates a new LRU cache with the specified maximum size.
// maxSize is the maximum number of entries to store.
func NewLRUCache(maxSize int) *LRUCache {
	if maxSize <= 0 {
		maxSize = 100 // Default
	}

	return &LRUCache{
		maxSize: maxSize,
		entries: make(map[string]*cacheEntry),
	}
}

// Get retrieves candles from cache. Returns nil if not found.
func (c *LRUCache) Get(key string) []*market.Candle {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		c.misses++
		return nil
	}

	c.hits++

	// Move to front (most recently used)
	c.moveToFront(entry)

	return entry.candles
}

// Put stores candles in cache.
func (c *LRUCache) Put(key string, candles []*market.Candle) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if entry, exists := c.entries[key]; exists {
		// Update existing entry
		entry.candles = candles
		entry.cachedAt = time.Now()
		c.moveToFront(entry)
		return
	}

	// Create new entry
	entry := &cacheEntry{
		key:      key,
		candles:  candles,
		cachedAt: time.Now(),
	}

	c.entries[key] = entry
	c.addToFront(entry)

	// Evict if over capacity
	if len(c.entries) > c.maxSize {
		c.evictLRU()
	}
}

// Clear removes all entries from cache.
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
	c.head = nil
	c.tail = nil
}

// Stats returns cache statistics.
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      len(c.entries),
		MaxSize:   c.maxSize,
		HitRate:   hitRate,
	}
}

// moveToFront moves an entry to the front of the LRU list.
func (c *LRUCache) moveToFront(entry *cacheEntry) {
	if entry == c.head {
		return // Already at front
	}

	// Remove from current position
	c.removeFromList(entry)

	// Add to front
	c.addToFront(entry)
}

// addToFront adds an entry to the front of the LRU list.
func (c *LRUCache) addToFront(entry *cacheEntry) {
	entry.next = c.head
	entry.prev = nil

	if c.head != nil {
		c.head.prev = entry
	}
	c.head = entry

	if c.tail == nil {
		c.tail = entry
	}
}

// removeFromList removes an entry from the LRU list.
func (c *LRUCache) removeFromList(entry *cacheEntry) {
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		c.head = entry.next
	}

	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		c.tail = entry.prev
	}

	entry.prev = nil
	entry.next = nil
}

// evictLRU removes the least recently used entry.
func (c *LRUCache) evictLRU() {
	if c.tail == nil {
		return
	}

	// Remove tail (least recently used)
	toEvict := c.tail
	c.removeFromList(toEvict)
	delete(c.entries, toEvict.key)
	c.evictions++
}

// GenerateCacheKey generates a cache key from symbol, timeframe, and candle data hash.
func GenerateCacheKey(symbol, timeframe string, candles []*market.Candle) string {
	// Include symbol and timeframe
	key := fmt.Sprintf("%s:%s:", symbol, timeframe)

	// Add hash of candle data for uniqueness
	hash := hashCandles(candles)
	key += hash

	return key
}

// hashCandles creates a hash of candle data for cache key generation.
// Uses first/last timestamps and count for lightweight hashing.
func hashCandles(candles []*market.Candle) string {
	if len(candles) == 0 {
		return "empty"
	}

	h := sha256.New()

	// Include count
	h.Write([]byte(fmt.Sprintf("count:%d", len(candles))))

	// Include first timestamp
	if len(candles) > 0 {
		h.Write([]byte(fmt.Sprintf("first:%d", candles[0].Timestamp.Unix())))
	}

	// Include last timestamp
	if len(candles) > 0 {
		h.Write([]byte(fmt.Sprintf("last:%d", candles[len(candles)-1].Timestamp.Unix())))
	}

	// Include first/last OHLC for additional uniqueness
	if len(candles) > 0 {
		c := candles[0]
		h.Write([]byte(fmt.Sprintf("ohlc:%f:%f:%f:%f", c.Open, c.High, c.Low, c.Close)))
	}
	if len(candles) > 1 {
		c := candles[len(candles)-1]
		h.Write([]byte(fmt.Sprintf("ohlc:%f:%f:%f:%f", c.Open, c.High, c.Low, c.Close)))
	}

	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars
}

// NoOpCache is a cache that does nothing (always miss).
// Useful for disabling caching without changing code.
type NoOpCache struct{}

// NewNoOpCache creates a no-op cache.
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns nil (cache miss).
func (c *NoOpCache) Get(key string) []*market.Candle {
	return nil
}

// Put does nothing.
func (c *NoOpCache) Put(key string, candles []*market.Candle) {
	// No-op
}

// Clear does nothing.
func (c *NoOpCache) Clear() {
	// No-op
}

// Stats returns empty stats.
func (c *NoOpCache) Stats() CacheStats {
	return CacheStats{}
}
