package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// BenchmarkLRUCache_Put benchmarks cache put operations
func BenchmarkLRUCache_Put(b *testing.B) {
	cache := NewLRUCache(1000)
	candles := generateTestCandles(100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		cache.Put(key, candles)
	}
}

// BenchmarkLRUCache_Get_Hit benchmarks cache hit operations
func BenchmarkLRUCache_Get_Hit(b *testing.B) {
	cache := NewLRUCache(1000)
	candles := generateTestCandles(100)
	
	// Populate cache
	for i := 0; i < 100; i++ {
		cache.Put(fmt.Sprintf("key-%d", i), candles)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key-%d", i%100))
	}
}

// BenchmarkLRUCache_Get_Miss benchmarks cache miss operations
func BenchmarkLRUCache_Get_Miss(b *testing.B) {
	cache := NewLRUCache(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("nonexistent-%d", i))
	}
}

// BenchmarkLRUCache_Mixed benchmarks mixed read/write operations
func BenchmarkLRUCache_Mixed(b *testing.B) {
	cache := NewLRUCache(1000)
	candles := generateTestCandles(100)
	
	// Populate with some data
	for i := 0; i < 100; i++ {
		cache.Put(fmt.Sprintf("key-%d", i), candles)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%3 == 0 {
			// Write
			cache.Put(fmt.Sprintf("key-%d", i%1000), candles)
		} else {
			// Read
			cache.Get(fmt.Sprintf("key-%d", i%100))
		}
	}
}

// BenchmarkLRUCache_Eviction benchmarks eviction behavior
func BenchmarkLRUCache_Eviction(b *testing.B) {
	cache := NewLRUCache(100) // Small cache to trigger evictions
	candles := generateTestCandles(100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		cache.Put(key, candles)
	}
}

// BenchmarkGenerateCacheKey benchmarks cache key generation
func BenchmarkGenerateCacheKey(b *testing.B) {
	candles := generateTestCandles(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCacheKey("BTC", "1h", candles)
	}
}

// BenchmarkGenerateCacheKey_Small benchmarks key generation with small dataset
func BenchmarkGenerateCacheKey_Small(b *testing.B) {
	candles := generateTestCandles(10)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCacheKey("BTC", "1h", candles)
	}
}

// BenchmarkGenerateCacheKey_Large benchmarks key generation with large dataset
func BenchmarkGenerateCacheKey_Large(b *testing.B) {
	candles := generateTestCandles(10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCacheKey("BTC", "1h", candles)
	}
}

// BenchmarkHashCandles benchmarks candle hashing
func BenchmarkHashCandles(b *testing.B) {
	candles := generateTestCandles(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashCandles(candles)
	}
}

// BenchmarkLRUCache_LargeDataset benchmarks cache with large candle datasets
func BenchmarkLRUCache_LargeDataset(b *testing.B) {
	cache := NewLRUCache(100)
	candles := generateTestCandles(10000) // 10K candles
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		cache.Put(key, candles)
	}
}

// BenchmarkLRUCache_Stats benchmarks stats retrieval
func BenchmarkLRUCache_Stats(b *testing.B) {
	cache := NewLRUCache(1000)
	candles := generateTestCandles(100)
	
	// Populate cache
	for i := 0; i < 100; i++ {
		cache.Put(fmt.Sprintf("key-%d", i), candles)
		cache.Get(fmt.Sprintf("key-%d", i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Stats()
	}
}

// BenchmarkNoOpCache benchmarks no-op cache
func BenchmarkNoOpCache(b *testing.B) {
	cache := NewNoOpCache()
	candles := generateTestCandles(100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		cache.Put(key, candles)
		cache.Get(key)
	}
}

// generateTestCandles generates test candle data
func generateTestCandles(count int) []*market.Candle {
	candles := make([]*market.Candle, count)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	basePrice := 100.0

	for i := 0; i < count; i++ {
		open := basePrice + float64(i%10)
		high := open + float64(i%5) + 1
		low := open - float64(i%3)
		close := open + float64((i%7)-3)
		volume := 1000.0 + float64(i%500)

		candles[i] = &market.Candle{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}
	}

	return candles
}
