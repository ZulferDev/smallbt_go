package indicator

import (
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// BenchmarkATR_Stateless benchmarks the original stateless ATR implementation.
func BenchmarkATR_Stateless(b *testing.B) {
	candles := generateTestCandles(1000)
	period := 14

	atr := &ATR{name: "atr", period: period}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range candles {
			ctx := &Context{
				Current:  candles[j],
				Candles:  candles[:j+1],
				BarIndex: j,
			}
			_, _ = atr.Calculate(ctx)
		}
	}
}

// BenchmarkATR_Cached benchmarks the optimized cached ATR implementation.
func BenchmarkATR_Cached(b *testing.B) {
	candles := generateTestCandles(1000)
	period := 14

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atr, _ := NewCachedATR(period)
		var prevCandle *market.Candle

		for _, candle := range candles {
			_, _ = atr.Update(candle, prevCandle)
			prevCandle = &candle
		}
	}
}

// BenchmarkATR_Stateless_5Years benchmarks stateless ATR with 5 years of data.
func BenchmarkATR_Stateless_5Years(b *testing.B) {
	candles := generateTestCandles(43800) // 5 years hourly
	period := 14

	atr := &ATR{name: "atr", period: period}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range candles {
			ctx := &Context{
				Current:  candles[j],
				Candles:  candles[:j+1],
				BarIndex: j,
			}
			_, _ = atr.Calculate(ctx)
		}
	}
}

// BenchmarkATR_Cached_5Years benchmarks cached ATR with 5 years of data.
func BenchmarkATR_Cached_5Years(b *testing.B) {
	candles := generateTestCandles(43800) // 5 years hourly
	period := 14

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atr, _ := NewCachedATR(period)
		var prevCandle *market.Candle

		for _, candle := range candles {
			_, _ = atr.Update(candle, prevCandle)
			prevCandle = &candle
		}
	}
}

// BenchmarkStateManager benchmarks StateManager with multiple indicators.
func BenchmarkStateManager(b *testing.B) {
	candles := generateTestCandles(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm := NewStateManager()

		atr14, _ := NewCachedATR(14)
		atr20, _ := NewCachedATR(20)

		sm.Register("atr14", atr14)
		sm.Register("atr20", atr20)

		var prevCandle *market.Candle
		for _, candle := range candles {
			_ = sm.Update(candle, prevCandle)
			prevCandle = &candle
		}
	}
}

// BenchmarkCachedRegistry_Create benchmarks indicator creation via registry.
func BenchmarkCachedRegistry_Create(b *testing.B) {
	reg := BuiltinCachedRegistry()
	config := Config{
		Type:   "atr",
		Period: 14,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reg.CreateCached(config)
	}
}

// BenchmarkStatelessWrapper benchmarks wrapped stateless indicators.
func BenchmarkStatelessWrapper(b *testing.B) {
	reg := BuiltinCachedRegistry()
	config := Config{
		Type:   "sma",
		Period: 20,
		Source: "close",
	}

	ind, _ := reg.CreateCached(config)
	candles := generateTestCandles(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind.Reset()
		var prevCandle *market.Candle
		for _, candle := range candles {
			_, _ = ind.Update(candle, prevCandle)
			prevCandle = &candle
		}
	}
}
