package transform

import (
	"context"
	"fmt"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/data/cache"
	"github.com/ZulferDev/smallbt_go/internal/data/csv"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// MultiSymbol benchmarks

func BenchmarkMultiSymbolTransformedFeed_Next(b *testing.B) {
	csvPath := createTestCSV(b, 10000)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")

	msf := NewMultiSymbolTransformedFeed()

	for _, count := range []int{1, 3, 5, 10} {
		b.Run(fmt.Sprintf("%d_symbols", count), func(b *testing.B) {
			// Setup
			for i := 0; i < count; i++ {
				feed, _ := csv.NewCSVDataFeed(csvPath, config)
				chain := NewTransformChain(NewScaleTransform(2.0, "close"))
				symbol := fmt.Sprintf("SYM%d", i)
				_ = msf.AddSymbol(symbol, feed, chain, 100)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = msf.Next("SYM0")
			}
		})
	}
}

func BenchmarkMultiSymbolTransformedFeed_NextAll(b *testing.B) {
	csvPath := createTestCSV(b, 1000)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")

	for _, count := range []int{3, 5, 10} {
		b.Run(fmt.Sprintf("%d_symbols", count), func(b *testing.B) {
			msf := NewMultiSymbolTransformedFeed()

			// Setup
			for i := 0; i < count; i++ {
				feed, _ := csv.NewCSVDataFeed(csvPath, config)
				chain := NewTransformChain(NewScaleTransform(2.0, "close"))
				symbol := fmt.Sprintf("SYM%d", i)
				_ = msf.AddSymbol(symbol, feed, chain, 100)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = msf.NextAll()
			}
		})
	}
}

func BenchmarkParallelMultiSymbolFeed(b *testing.B) {
	csvPath := createTestCSV(b, 5000)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")

	for _, workers := range []int{1, 2, 4, 8} {
		for _, symbols := range []int{5, 10, 20} {
			b.Run(fmt.Sprintf("workers_%d_symbols_%d", workers, symbols), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					pmsf := NewParallelMultiSymbolFeed(workers)

					for j := 0; j < symbols; j++ {
						feed, _ := csv.NewCSVDataFeed(csvPath, config)
						chain := NewTransformChain(NewScaleTransform(2.0, "close"))
						_ = pmsf.AddSymbol(fmt.Sprintf("SYM%d", j), feed, chain, 100)
					}

					ctx := context.Background()
					b.StartTimer()

					_, _ = pmsf.ReadAllParallel(ctx)

					b.StopTimer()
					pmsf.Close()
				}
			})
		}
	}
}

// Cache benchmarks

func BenchmarkCachedTransformedFeed_ReadAll(b *testing.B) {
	csvPath := createTestCSV(b, 10000)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")

	for _, cacheSize := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("cache_%d", cacheSize), func(b *testing.B) {
			lruCache := cache.NewLRUCache(cacheSize)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				feed, _ := csv.NewCSVDataFeed(csvPath, config)
				chain := NewTransformChain(NewScaleTransform(2.0, "close"))
				ctf, _ := NewCachedTransformedFeed(feed, chain, lruCache, "test_key", 100)
				b.StartTimer()

				// First read (miss)
				_, _ = ctf.ReadAll()
				// Second read (hit)
				_, _ = ctf.ReadAll()

				b.StopTimer()
				ctf.Close()
			}
		})
	}
}

func BenchmarkSmartCache_GetOrCompute(b *testing.B) {
	candles := generateTestCandles(1000, 100.0)

	for _, threshold := range []int{1, 2, 5, 10} {
		b.Run(fmt.Sprintf("threshold_%d", threshold), func(b *testing.B) {
			sc := NewSmartCache(100, threshold)

			compute := func() ([]*market.Candle, error) {
				return candles, nil
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("key_%d", i%10) // Reuse 10 keys
				_, _ = sc.GetOrCompute(key, compute)
			}
		})
	}
}

func BenchmarkTransformCache_Operations(b *testing.B) {
	tc := NewTransformCache(1000)
	candles := generateTestCandles(1000, 100.0)
	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	b.Run("Put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := tc.GenerateKey("BTC", "1h", chain)
			tc.Put(key, candles)
		}
	})

	b.Run("Get_Hit", func(b *testing.B) {
		key := tc.GenerateKey("BTC", "1h", chain)
		tc.Put(key, candles)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tc.Get(key)
		}
	})

	b.Run("Get_Miss", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tc.Get("nonexistent_key")
		}
	})
}

// Conditional transform benchmarks

func BenchmarkConditionalTransform_Apply(b *testing.B) {
	candles := generateTestCandles(10000, 100.0)
	transform := NewScaleTransform(2.0, "close")

	conditions := map[string]Condition{
		"VolumeAbove": VolumeAbove(50000.0),
		"PriceAbove":  PriceAbove(100.0),
		"Bullish":     BullishCandle(),
		"Complex":     AndCondition(VolumeAbove(50000.0), BullishCandle()),
	}

	for name, condition := range conditions {
		b.Run(name, func(b *testing.B) {
			ct := NewConditionalTransform(transform, condition)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ct.Apply(candles)
			}
		})
	}
}

func BenchmarkSwitchTransform_Apply(b *testing.B) {
	candles := generateTestCandles(10000, 100.0)

	for _, numCases := range []int{2, 5, 10} {
		b.Run(fmt.Sprintf("%d_cases", numCases), func(b *testing.B) {
			st := NewSwitchTransform(NewScaleTransform(1.0, "close"))

			for i := 0; i < numCases; i++ {
				threshold := 100.0 + float64(i*10)
				st.AddCase(PriceAbove(threshold), NewScaleTransform(float64(i+1), "close"))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = st.Apply(candles)
			}
		})
	}
}

func BenchmarkFilterTransform_Apply(b *testing.B) {
	candles := generateTestCandles(10000, 100.0)

	filters := map[string]Condition{
		"10pct": VolumeAbove(90000.0), // ~10% pass
		"50pct": VolumeAbove(50000.0), // ~50% pass
		"90pct": VolumeAbove(10000.0), // ~90% pass
	}

	for name, condition := range filters {
		b.Run(name, func(b *testing.B) {
			ft := NewFilterTransform(condition)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ft.Apply(candles)
			}
		})
	}
}

func BenchmarkConditionBuilders(b *testing.B) {
	candle := &market.Candle{
		Open:   100.0,
		High:   110.0,
		Low:    95.0,
		Close:  105.0,
		Volume: 100000.0,
	}

	conditions := map[string]Condition{
		"VolumeAbove":   VolumeAbove(50000.0),
		"PriceInRange":  PriceInRange(100.0, 110.0),
		"BullishCandle": BullishCandle(),
		"AndCondition":  AndCondition(PriceAbove(100.0), VolumeAbove(50000.0)),
		"OrCondition":   OrCondition(PriceAbove(200.0), VolumeAbove(50000.0)),
		"ComplexAnd":    AndCondition(PriceAbove(100.0), VolumeAbove(50000.0), BullishCandle()),
	}

	for name, condition := range conditions {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				condition(candle)
			}
		})
	}
}

// Integration benchmarks

func BenchmarkEndToEnd_MultiSymbol_Cached_Conditional(b *testing.B) {
	csvPath := createTestCSV(b, 5000)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		// Setup multi-symbol feed
		msf := NewMultiSymbolTransformedFeed()
		lruCache := cache.NewLRUCache(10)

		for j := 0; j < 3; j++ {
			feed, _ := csv.NewCSVDataFeed(csvPath, config)

			// Conditional transform: scale high volume candles
			transform := NewScaleTransform(2.0, "close")
			condition := VolumeAbove(50000.0)
			ct := NewConditionalTransform(transform, condition)

			chain := NewTransformChain(ct)

			// Cached transformed feed
			ctf, _ := NewCachedTransformedFeed(feed, chain, lruCache, fmt.Sprintf("sym%d", j), 100)

			symbol := fmt.Sprintf("SYM%d", j)
			msf.feeds[symbol] = &TransformedFeed{
				feed:      ctf.feed.feed,
				chain:     chain,
				batchSize: 100,
			}
		}

		b.StartTimer()

		// Read all symbols
		for j := 0; j < 3; j++ {
			msf.ReadAll(fmt.Sprintf("SYM%d", j))
		}

		b.StopTimer()
		msf.Close()
	}
}
