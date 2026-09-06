package resample

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// BenchmarkResampler_1mTo5m benchmarks 1m → 5m resampling
func BenchmarkResampler_1mTo5m(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Generate 1000 x 1-minute candles
	candles := generateCandles(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe5m)
	}
}

// BenchmarkResampler_1mTo15m benchmarks 1m → 15m resampling
func BenchmarkResampler_1mTo15m(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Generate 1000 x 1-minute candles
	candles := generateCandles(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe15m)
	}
}

// BenchmarkResampler_1mTo1h benchmarks 1m → 1h resampling
func BenchmarkResampler_1mTo1h(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Generate 1000 x 1-minute candles
	candles := generateCandles(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe1h)
	}
}

// BenchmarkResampler_1mTo4h benchmarks 1m → 4h resampling
func BenchmarkResampler_1mTo4h(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Generate 1000 x 1-minute candles
	candles := generateCandles(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe4h)
	}
}

// BenchmarkResampler_1mTo1d benchmarks 1m → 1d resampling
func BenchmarkResampler_1mTo1d(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Generate 1440 x 1-minute candles (1 day)
	candles := generateCandles(1440, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe1d)
	}
}

// BenchmarkResampler_5mTo1h benchmarks 5m → 1h resampling
func BenchmarkResampler_5mTo1h(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe5m)

	// Generate 1000 x 5-minute candles
	candles := generateCandles(1000, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe1h)
	}
}

// BenchmarkResampler_1hTo4h benchmarks 1h → 4h resampling
func BenchmarkResampler_1hTo4h(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1h)

	// Generate 1000 x 1-hour candles
	candles := generateCandles(1000, time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe4h)
	}
}

// BenchmarkResampler_1hTo1d benchmarks 1h → 1d resampling
func BenchmarkResampler_1hTo1d(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1h)

	// Generate 1000 x 1-hour candles
	candles := generateCandles(1000, time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe1d)
	}
}

// BenchmarkResampler_LargeDataset_1mTo5m benchmarks large dataset (10K candles)
func BenchmarkResampler_LargeDataset_1mTo5m(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Generate 10,000 x 1-minute candles
	candles := generateCandles(10000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe5m)
	}
}

// BenchmarkResampler_LargeDataset_1mTo1h benchmarks large dataset (10K candles)
func BenchmarkResampler_LargeDataset_1mTo1h(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Generate 10,000 x 1-minute candles
	candles := generateCandles(10000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe1h)
	}
}

// BenchmarkResampler_LargeDataset_1mTo1d benchmarks large dataset (10K candles)
func BenchmarkResampler_LargeDataset_1mTo1d(b *testing.B) {
	r := NewDefaultResampler(market.Timeframe1m)

	// Generate 10,000 x 1-minute candles
	candles := generateCandles(10000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Resample(candles, market.Timeframe1d)
	}
}

// BenchmarkAlignTimestamp benchmarks timestamp alignment
func BenchmarkAlignTimestamp(b *testing.B) {
	t := time.Date(2024, 1, 1, 14, 23, 45, 0, time.UTC)
	period := 5 * time.Minute

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AlignTimestamp(t, period)
	}
}

// BenchmarkParseTimeframe benchmarks timeframe parsing
func BenchmarkParseTimeframe(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseTimeframe(market.Timeframe1m)
	}
}

// generateCandles generates test candles with realistic OHLCV data
func generateCandles(count int, interval time.Duration) []*market.Candle {
	candles := make([]*market.Candle, count)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	basePrice := 100.0

	for i := 0; i < count; i++ {
		// Simulate price movement
		open := basePrice + float64(i%10)
		high := open + float64(i%5) + 1
		low := open - float64(i%3)
		close := open + float64((i%7)-3)
		volume := 1000.0 + float64(i%500)

		candles[i] = &market.Candle{
			Timestamp: baseTime.Add(time.Duration(i) * interval),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}
	}

	return candles
}
