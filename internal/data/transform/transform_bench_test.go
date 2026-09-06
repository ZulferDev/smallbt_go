package transform

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Benchmark helpers

func generateBenchCandles(count int) []*market.Candle {
	candles := make([]*market.Candle, count)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		price := 100.0 + float64(i)*0.1
		candles[i] = &market.Candle{
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour),
			Open:      price,
			High:      price + 1.0,
			Low:       price - 0.5,
			Close:     price + 0.5,
			Volume:    1000.0 + float64(i),
		}
	}

	return candles
}

// Normalize benchmarks

func BenchmarkNormalize_100(b *testing.B) {
	candles := generateBenchCandles(100)
	transform := NewNormalizeTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkNormalize_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewNormalizeTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkNormalize_10K(b *testing.B) {
	candles := generateBenchCandles(10000)
	transform := NewNormalizeTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkNormalize_AllFields(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewNormalizeTransform("all")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

// LogReturns benchmarks

func BenchmarkLogReturns_100(b *testing.B) {
	candles := generateBenchCandles(100)
	transform := NewLogReturnsTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkLogReturns_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewLogReturnsTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkLogReturns_10K(b *testing.B) {
	candles := generateBenchCandles(10000)
	transform := NewLogReturnsTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

// PercentageChange benchmarks

func BenchmarkPercentageChange_100(b *testing.B) {
	candles := generateBenchCandles(100)
	transform := NewPercentageChangeTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkPercentageChange_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewPercentageChangeTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkPercentageChange_10K(b *testing.B) {
	candles := generateBenchCandles(10000)
	transform := NewPercentageChangeTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

// MovingAverageSmooth benchmarks

func BenchmarkMASmooth_Period5_100(b *testing.B) {
	candles := generateBenchCandles(100)
	transform := NewMovingAverageSmoothTransform(5, "close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkMASmooth_Period5_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewMovingAverageSmoothTransform(5, "close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkMASmooth_Period20_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewMovingAverageSmoothTransform(20, "close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkMASmooth_Period50_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewMovingAverageSmoothTransform(50, "close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

// Difference benchmarks

func BenchmarkDifference_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewDifferenceTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkDifference_10K(b *testing.B) {
	candles := generateBenchCandles(10000)
	transform := NewDifferenceTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

// Scale benchmarks

func BenchmarkScale_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewScaleTransform(2.0, "close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkScale_AllFields_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewScaleTransform(2.0, "all")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

// TransformChain benchmarks

func BenchmarkChain_TwoTransforms_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewScaleTransform(100.0, "close"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chain.Apply(candles)
	}
}

func BenchmarkChain_ThreeTransforms_1K(b *testing.B) {
	candles := generateBenchCandles(1000)
	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewMovingAverageSmoothTransform(5, "close"),
		NewScaleTransform(100.0, "close"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chain.Apply(candles)
	}
}

func BenchmarkChain_Complex_1K(b *testing.B) {
	candles := generateBenchCandles(1001) // +1 for difference
	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewMovingAverageSmoothTransform(10, "close"),
		NewDifferenceTransform("close"),
		NewScaleTransform(100.0, "close"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chain.Apply(candles)
	}
}

// Helper function benchmarks

func BenchmarkValidateCandles_1K(b *testing.B) {
	candles := generateBenchCandles(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateCandles(candles)
	}
}

func BenchmarkCopyCandles_100(b *testing.B) {
	candles := generateBenchCandles(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CopyCandles(candles)
	}
}

func BenchmarkCopyCandles_1K(b *testing.B) {
	candles := generateBenchCandles(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CopyCandles(candles)
	}
}

func BenchmarkCopyCandles_10K(b *testing.B) {
	candles := generateBenchCandles(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CopyCandles(candles)
	}
}

// Memory allocation benchmarks

func BenchmarkNormalize_1K_Allocs(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewNormalizeTransform("close")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkLogReturns_1K_Allocs(b *testing.B) {
	candles := generateBenchCandles(1000)
	transform := NewLogReturnsTransform("close")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkChain_1K_Allocs(b *testing.B) {
	candles := generateBenchCandles(1000)
	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewScaleTransform(100.0, "close"),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chain.Apply(candles)
	}
}

// Large dataset benchmarks

func BenchmarkNormalize_100K(b *testing.B) {
	candles := generateBenchCandles(100000)
	transform := NewNormalizeTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkLogReturns_100K(b *testing.B) {
	candles := generateBenchCandles(100000)
	transform := NewLogReturnsTransform("close")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transform.Apply(candles)
	}
}

func BenchmarkChain_100K(b *testing.B) {
	candles := generateBenchCandles(100000)
	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewMovingAverageSmoothTransform(20, "close"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chain.Apply(candles)
	}
}
