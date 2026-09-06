package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache(10)
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache.maxSize != 10 {
		t.Errorf("expected maxSize 10, got %d", cache.maxSize)
	}
	
	stats := cache.Stats()
	if stats.MaxSize != 10 {
		t.Errorf("expected MaxSize 10, got %d", stats.MaxSize)
	}
	if stats.Size != 0 {
		t.Errorf("expected initial size 0, got %d", stats.Size)
	}
}

func TestNewLRUCache_DefaultSize(t *testing.T) {
	cache := NewLRUCache(0)
	if cache.maxSize != 100 {
		t.Errorf("expected default maxSize 100, got %d", cache.maxSize)
	}
}

func TestLRUCache_PutAndGet(t *testing.T) {
	cache := NewLRUCache(10)
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	key := "BTC:1h:test123"
	
	// Put
	cache.Put(key, candles)
	
	// Get
	retrieved := cache.Get(key)
	if retrieved == nil {
		t.Fatal("expected to retrieve candles, got nil")
	}
	if len(retrieved) != 1 {
		t.Errorf("expected 1 candle, got %d", len(retrieved))
	}
	if retrieved[0].Open != 100 {
		t.Errorf("expected open 100, got %f", retrieved[0].Open)
	}
}

func TestLRUCache_GetMiss(t *testing.T) {
	cache := NewLRUCache(10)
	
	retrieved := cache.Get("nonexistent")
	if retrieved != nil {
		t.Error("expected nil for cache miss")
	}
	
	stats := cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestLRUCache_Stats(t *testing.T) {
	cache := NewLRUCache(10)
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	// Put 3 items
	cache.Put("key1", candles)
	cache.Put("key2", candles)
	cache.Put("key3", candles)
	
	// Get key1 (hit)
	cache.Get("key1")
	
	// Get key2 (hit)
	cache.Get("key2")
	
	// Get nonexistent (miss)
	cache.Get("keyX")
	
	stats := cache.Stats()
	
	if stats.Hits != 2 {
		t.Errorf("expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.Size != 3 {
		t.Errorf("expected size 3, got %d", stats.Size)
	}
	
	expectedHitRate := 2.0 / 3.0
	if stats.HitRate != expectedHitRate {
		t.Errorf("expected hit rate %f, got %f", expectedHitRate, stats.HitRate)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(3) // Small cache
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	// Fill cache to capacity
	cache.Put("key1", candles)
	cache.Put("key2", candles)
	cache.Put("key3", candles)
	
	stats := cache.Stats()
	if stats.Size != 3 {
		t.Errorf("expected size 3, got %d", stats.Size)
	}
	
	// Add one more (should evict key1, the least recently used)
	cache.Put("key4", candles)
	
	stats = cache.Stats()
	if stats.Size != 3 {
		t.Errorf("expected size still 3 after eviction, got %d", stats.Size)
	}
	if stats.Evictions != 1 {
		t.Errorf("expected 1 eviction, got %d", stats.Evictions)
	}
	
	// key1 should be evicted
	if cache.Get("key1") != nil {
		t.Error("expected key1 to be evicted")
	}
	
	// key2, key3, key4 should still exist
	if cache.Get("key2") == nil {
		t.Error("expected key2 to exist")
	}
	if cache.Get("key3") == nil {
		t.Error("expected key3 to exist")
	}
	if cache.Get("key4") == nil {
		t.Error("expected key4 to exist")
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache(3)
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	// Fill cache
	cache.Put("key1", candles)
	cache.Put("key2", candles)
	cache.Put("key3", candles)
	
	// Access key1 (move to front)
	cache.Get("key1")
	
	// Add key4 (should evict key2, not key1)
	cache.Put("key4", candles)
	
	// key1 should still exist (recently accessed)
	if cache.Get("key1") == nil {
		t.Error("expected key1 to exist (recently accessed)")
	}
	
	// key2 should be evicted (least recently used)
	if cache.Get("key2") != nil {
		t.Error("expected key2 to be evicted")
	}
	
	// key3 and key4 should exist
	if cache.Get("key3") == nil {
		t.Error("expected key3 to exist")
	}
	if cache.Get("key4") == nil {
		t.Error("expected key4 to exist")
	}
}

func TestLRUCache_UpdateExisting(t *testing.T) {
	cache := NewLRUCache(10)
	
	candles1 := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	candles2 := []*market.Candle{
		{Timestamp: time.Now(), Open: 200, High: 210, Low: 190, Close: 205, Volume: 2000},
	}
	
	key := "BTC:1h:test"
	
	// Put initial
	cache.Put(key, candles1)
	
	stats := cache.Stats()
	if stats.Size != 1 {
		t.Errorf("expected size 1, got %d", stats.Size)
	}
	
	// Update with new data
	cache.Put(key, candles2)
	
	stats = cache.Stats()
	if stats.Size != 1 {
		t.Errorf("expected size still 1 after update, got %d", stats.Size)
	}
	
	// Retrieve should get updated data
	retrieved := cache.Get(key)
	if retrieved[0].Open != 200 {
		t.Errorf("expected updated open 200, got %f", retrieved[0].Open)
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10)
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	cache.Put("key1", candles)
	cache.Put("key2", candles)
	cache.Put("key3", candles)
	
	stats := cache.Stats()
	if stats.Size != 3 {
		t.Errorf("expected size 3, got %d", stats.Size)
	}
	
	cache.Clear()
	
	stats = cache.Stats()
	if stats.Size != 0 {
		t.Errorf("expected size 0 after clear, got %d", stats.Size)
	}
	
	if cache.Get("key1") != nil {
		t.Error("expected key1 to be cleared")
	}
}

func TestGenerateCacheKey(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	candles1 := []*market.Candle{
		{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: t0.Add(1 * time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
	}
	
	key1 := GenerateCacheKey("BTC", "1h", candles1)
	
	if key1 == "" {
		t.Error("expected non-empty key")
	}
	
	// Same data should produce same key
	key2 := GenerateCacheKey("BTC", "1h", candles1)
	if key1 != key2 {
		t.Error("same data should produce same key")
	}
	
	// Different symbol should produce different key
	key3 := GenerateCacheKey("ETH", "1h", candles1)
	if key1 == key3 {
		t.Error("different symbol should produce different key")
	}
	
	// Different timeframe should produce different key
	key4 := GenerateCacheKey("BTC", "4h", candles1)
	if key1 == key4 {
		t.Error("different timeframe should produce different key")
	}
}

func TestGenerateCacheKey_DifferentData(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	candles1 := []*market.Candle{
		{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	candles2 := []*market.Candle{
		{Timestamp: t0, Open: 200, High: 210, Low: 190, Close: 205, Volume: 2000},
	}
	
	key1 := GenerateCacheKey("BTC", "1h", candles1)
	key2 := GenerateCacheKey("BTC", "1h", candles2)
	
	// Different data should produce different key (OHLC differs)
	if key1 == key2 {
		t.Error("different data should produce different key")
	}
}

func TestGenerateCacheKey_EmptyCandles(t *testing.T) {
	key := GenerateCacheKey("BTC", "1h", []*market.Candle{})
	
	if key == "" {
		t.Error("expected non-empty key for empty candles")
	}
	
	// Should contain "empty" marker
	if len(key) < 10 {
		t.Error("expected reasonable key length")
	}
}

func TestHashCandles(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	candles := []*market.Candle{
		{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: t0.Add(1 * time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
	}
	
	hash := hashCandles(candles)
	
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if len(hash) != 16 {
		t.Errorf("expected hash length 16, got %d", len(hash))
	}
	
	// Same candles should produce same hash
	hash2 := hashCandles(candles)
	if hash != hash2 {
		t.Error("same candles should produce same hash")
	}
}

func TestHashCandles_Empty(t *testing.T) {
	hash := hashCandles([]*market.Candle{})
	if hash != "empty" {
		t.Errorf("expected 'empty' for empty candles, got %s", hash)
	}
}

func TestNoOpCache(t *testing.T) {
	cache := NewNoOpCache()
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	// Put should not error
	cache.Put("key", candles)
	
	// Get should always return nil
	retrieved := cache.Get("key")
	if retrieved != nil {
		t.Error("NoOpCache should always return nil")
	}
	
	// Clear should not error
	cache.Clear()
	
	// Stats should be empty
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("NoOpCache stats should be empty")
	}
}

func TestLRUCache_Concurrent(t *testing.T) {
	cache := NewLRUCache(100)
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	// Concurrent puts
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("key-%d-%d", n, j)
				cache.Put(key, candles)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	stats := cache.Stats()
	if stats.Size == 0 {
		t.Error("expected non-zero cache size after concurrent puts")
	}
}

func TestLRUCache_ConcurrentReadWrite(t *testing.T) {
	cache := NewLRUCache(50)
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	
	// Populate cache
	for i := 0; i < 10; i++ {
		cache.Put(fmt.Sprintf("key%d", i), candles)
	}
	
	// Concurrent reads and writes
	done := make(chan bool, 20)
	
	// 10 readers
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				cache.Get(fmt.Sprintf("key%d", n%10))
			}
			done <- true
		}(i)
	}
	
	// 10 writers
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				cache.Put(fmt.Sprintf("key%d", n%10), candles)
			}
			done <- true
		}(i)
	}
	
	// Wait for all
	for i := 0; i < 20; i++ {
		<-done
	}
	
	// Should not panic or deadlock
	stats := cache.Stats()
	if stats.Size == 0 {
		t.Error("expected non-zero cache size")
	}
}
