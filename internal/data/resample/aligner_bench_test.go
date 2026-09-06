package resample

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// BenchmarkAligner_TwoSymbols benchmarks alignment of 2 symbols with 1000 candles each
func BenchmarkAligner_TwoSymbols(b *testing.B) {
	a := NewDefaultAligner()

	// Generate 1000 candles for each symbol
	btc := generateCandles(1000, time.Minute)
	eth := generateCandles(1000, time.Minute)

	input := map[string][]*market.Candle{
		"BTC": btc,
		"ETH": eth,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Align(input)
	}
}

// BenchmarkAligner_FiveSymbols benchmarks alignment of 5 symbols with 1000 candles each
func BenchmarkAligner_FiveSymbols(b *testing.B) {
	a := NewDefaultAligner()

	// Generate 1000 candles for each symbol
	candles := generateCandles(1000, time.Minute)

	input := map[string][]*market.Candle{
		"BTC":  candles,
		"ETH":  candles,
		"BNB":  candles,
		"SOL":  candles,
		"DOGE": candles,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Align(input)
	}
}

// BenchmarkAligner_TenSymbols benchmarks alignment of 10 symbols with 1000 candles each
func BenchmarkAligner_TenSymbols(b *testing.B) {
	a := NewDefaultAligner()

	// Generate 1000 candles for each symbol
	candles := generateCandles(1000, time.Minute)

	input := map[string][]*market.Candle{
		"BTC":  candles,
		"ETH":  candles,
		"BNB":  candles,
		"SOL":  candles,
		"DOGE": candles,
		"ADA":  candles,
		"XRP":  candles,
		"DOT":  candles,
		"MATIC": candles,
		"AVAX": candles,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Align(input)
	}
}

// BenchmarkAligner_ForwardFill benchmarks alignment with forward-fill (20% missing data)
func BenchmarkAligner_ForwardFill(b *testing.B) {
	a := NewDefaultAligner()
	a.FillStrategy = FillStrategyForward

	// Generate 1000 candles for BTC (complete)
	btc := generateCandles(1000, time.Minute)

	// Generate 800 candles for ETH (80% of BTC, 20% gaps)
	eth := generateCandles(800, time.Minute)

	input := map[string][]*market.Candle{
		"BTC": btc,
		"ETH": eth,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Align(input)
	}
}

// BenchmarkAligner_DropStrategy benchmarks alignment with drop strategy
func BenchmarkAligner_DropStrategy(b *testing.B) {
	a := NewDefaultAligner()
	a.FillStrategy = FillStrategyDrop

	// Generate 1000 candles for BTC (complete)
	btc := generateCandles(1000, time.Minute)

	// Generate 800 candles for ETH (80% of BTC)
	eth := generateCandles(800, time.Minute)

	input := map[string][]*market.Candle{
		"BTC": btc,
		"ETH": eth,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Align(input)
	}
}

// BenchmarkAligner_AlignToReference benchmarks reference-based alignment
func BenchmarkAligner_AlignToReference(b *testing.B) {
	a := NewDefaultAligner()

	// Generate 1000 candles for BTC (reference)
	btc := generateCandles(1000, time.Minute)

	// Generate 1200 candles for ETH (longer history)
	eth := generateCandles(1200, time.Minute)

	input := map[string][]*market.Candle{
		"BTC": btc,
		"ETH": eth,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.AlignToReference("BTC", input)
	}
}

// BenchmarkAligner_LargeDataset_TwoSymbols benchmarks 10K candles per symbol
func BenchmarkAligner_LargeDataset_TwoSymbols(b *testing.B) {
	a := NewDefaultAligner()

	// Generate 10,000 candles for each symbol
	btc := generateCandles(10000, time.Minute)
	eth := generateCandles(10000, time.Minute)

	input := map[string][]*market.Candle{
		"BTC": btc,
		"ETH": eth,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Align(input)
	}
}

// BenchmarkAligner_LargeDataset_FiveSymbols benchmarks 10K candles per symbol
func BenchmarkAligner_LargeDataset_FiveSymbols(b *testing.B) {
	a := NewDefaultAligner()

	// Generate 10,000 candles for each symbol
	candles := generateCandles(10000, time.Minute)

	input := map[string][]*market.Candle{
		"BTC":  candles,
		"ETH":  candles,
		"BNB":  candles,
		"SOL":  candles,
		"DOGE": candles,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = a.Align(input)
	}
}

// BenchmarkGetCommonTimeRange benchmarks common time range calculation
func BenchmarkGetCommonTimeRange(b *testing.B) {
	// Generate candles with different start/end times
	btc := generateCandles(1000, time.Minute)
	eth := generateCandlesWithOffset(1000, time.Minute, 100*time.Minute)

	input := map[string][]*market.Candle{
		"BTC": btc,
		"ETH": eth,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = GetCommonTimeRange(input)
	}
}

// BenchmarkFilterByTimeRange benchmarks time range filtering
func BenchmarkFilterByTimeRange(b *testing.B) {
	candles := generateCandles(1000, time.Minute)
	start := candles[200].Timestamp
	end := candles[800].Timestamp

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterByTimeRange(candles, start, end)
	}
}

// BenchmarkSortTimestamps benchmarks timestamp sorting
func BenchmarkSortTimestamps(b *testing.B) {
	// Generate unsorted timestamps
	timestamps := make([]time.Time, 1000)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 1000; i++ {
		// Reverse order for worst-case
		timestamps[i] = baseTime.Add(time.Duration(999-i) * time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy for each iteration
		ts := make([]time.Time, len(timestamps))
		copy(ts, timestamps)
		sortTimestamps(ts)
	}
}

// generateCandlesWithOffset generates candles with a time offset
func generateCandlesWithOffset(count int, interval, offset time.Duration) []*market.Candle {
	candles := make([]*market.Candle, count)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Add(offset)
	basePrice := 100.0

	for i := 0; i < count; i++ {
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
