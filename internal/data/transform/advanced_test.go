package transform

import (
	"context"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/data/cache"
	"github.com/ZulferDev/smallbt_go/internal/data/csv"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// MultiSymbol tests

func TestNewMultiSymbolTransformedFeed(t *testing.T) {
	msf := NewMultiSymbolTransformedFeed()
	if msf == nil {
		t.Error("Expected non-nil multi-symbol feed")
	}

	if len(msf.Symbols()) != 0 {
		t.Errorf("Expected 0 symbols, got %d", len(msf.Symbols()))
	}
}

func TestMultiSymbolTransformedFeed_AddSymbol(t *testing.T) {
	csvPath := createTestCSV(t, 10)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")
	feed, _ := csv.NewCSVDataFeed(csvPath, config)

	msf := NewMultiSymbolTransformedFeed()
	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	err := msf.AddSymbol("BTC", feed, chain, 5)
	if err != nil {
		t.Fatalf("AddSymbol failed: %v", err)
	}

	symbols := msf.Symbols()
	if len(symbols) != 1 {
		t.Errorf("Expected 1 symbol, got %d", len(symbols))
	}

	if symbols[0] != "BTC" {
		t.Errorf("Expected BTC, got %s", symbols[0])
	}
}

func TestMultiSymbolTransformedFeed_AddDuplicate(t *testing.T) {
	csvPath := createTestCSV(t, 10)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")
	feed1, _ := csv.NewCSVDataFeed(csvPath, config)
	feed2, _ := csv.NewCSVDataFeed(csvPath, config)

	msf := NewMultiSymbolTransformedFeed()
	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	msf.AddSymbol("BTC", feed1, chain, 5)
	err := msf.AddSymbol("BTC", feed2, chain, 5)
	
	if err == nil {
		t.Error("Expected error when adding duplicate symbol")
	}
}

func TestMultiSymbolTransformedFeed_Next(t *testing.T) {
	csvPath := createTestCSV(t, 50)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")
	feed, _ := csv.NewCSVDataFeed(csvPath, config)

	msf := NewMultiSymbolTransformedFeed()
	chain := NewTransformChain(NewScaleTransform(2.0, "close"))
	msf.AddSymbol("BTC", feed, chain, 10)

	candle, err := msf.Next("BTC")
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}

	expected := 202.0 // 101 * 2
	if !almostEqual(candle.Close, expected, 0.01) {
		t.Errorf("Expected %.2f, got %.2f", expected, candle.Close)
	}
}

func TestMultiSymbolTransformedFeed_ReadAll(t *testing.T) {
	csvPath := createTestCSV(t, 30)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")
	feed, _ := csv.NewCSVDataFeed(csvPath, config)

	msf := NewMultiSymbolTransformedFeed()
	chain := NewTransformChain(NewNormalizeTransform("close"))
	msf.AddSymbol("BTC", feed, chain, 10)

	candles, err := msf.ReadAll("BTC")
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(candles) != 30 {
		t.Errorf("Expected 30 candles, got %d", len(candles))
	}
}

// Cache tests

func TestCachedTransformedFeed(t *testing.T) {
	csvPath := createTestCSV(t, 50)
	config := csv.DefaultCSVConfig(market.Symbol("BTC"), "1h")
	feed, _ := csv.NewCSVDataFeed(csvPath, config)

	lruCache := cache.NewLRUCache(10)
	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	ctf, err := NewCachedTransformedFeed(feed, chain, lruCache, "BTC:1h:scale", 10)
	if err != nil {
		t.Fatalf("NewCachedTransformedFeed failed: %v", err)
	}
	defer ctf.Close()

	// First read - cache miss
	candles1, err := ctf.ReadAll()
	if err != nil {
		t.Fatalf("First ReadAll failed: %v", err)
	}

	if len(candles1) != 50 {
		t.Errorf("Expected 50 candles, got %d", len(candles1))
	}

	// Second read - cache hit
	candles2, err := ctf.ReadAll()
	if err != nil {
		t.Fatalf("Second ReadAll failed: %v", err)
	}

	if len(candles2) != 50 {
		t.Errorf("Expected 50 candles, got %d", len(candles2))
	}

	// Should be same data
	if candles1[0].Close != candles2[0].Close {
		t.Error("Cached data differs from original")
	}
}

func TestTransformCache(t *testing.T) {
	tc := NewTransformCache(10)
	chain := NewTransformChain(NewScaleTransform(2.0, "close"))

	key := tc.GenerateKey("BTC", "1h", chain)
	if key == "" {
		t.Error("Generated key is empty")
	}

	// Put and get
	candles := generateTestCandles(10, 100.0)
	tc.Put(key, candles)

	cached, ok := tc.Get(key)
	if !ok {
		t.Error("Cache miss on get")
	}

	if len(cached) != len(candles) {
		t.Errorf("Cached length %d != original %d", len(cached), len(candles))
	}
}

func TestSmartCache(t *testing.T) {
	sc := NewSmartCache(10, 2) // Cache after 2 accesses

	compute := func() ([]*market.Candle, error) {
		return generateTestCandles(10, 100.0), nil
	}

	// First access - not cached
	candles1, err := sc.GetOrCompute("key1", compute)
	if err != nil {
		t.Fatalf("First GetOrCompute failed: %v", err)
	}

	// Check not in cache yet
	if _, ok := sc.cache.Get("key1"); ok {
		t.Error("Should not be cached after 1 access")
	}

	// Second access - now cached
	candles2, err := sc.GetOrCompute("key1", compute)
	if err != nil {
		t.Fatalf("Second GetOrCompute failed: %v", err)
	}

	// Should be in cache now
	if _, ok := sc.cache.Get("key1"); !ok {
		t.Error("Should be cached after 2 accesses")
	}

	if len(candles1) != len(candles2) {
		t.Error("Candle count mismatch")
	}
}

// Conditional transform tests

func TestConditionalTransform(t *testing.T) {
	candles := []*market.Candle{
		{Close: 100.0, Volume: 1000.0},
		{Close: 105.0, Volume: 2000.0},
		{Close: 110.0, Volume: 1500.0},
		{Close: 115.0, Volume: 3000.0},
	}

	// Scale only high volume candles (>1800)
	condition := VolumeAbove(1800.0)
	transform := NewScaleTransform(2.0, "close")
	ct := NewConditionalTransform(transform, condition)

	result, err := ct.Apply(candles)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Check: candles with volume > 1800 should be scaled
	if !almostEqual(result[0].Close, 100.0, 0.01) {
		t.Errorf("Low volume candle should not be scaled: %.2f", result[0].Close)
	}

	if !almostEqual(result[1].Close, 210.0, 0.01) {
		t.Errorf("High volume candle should be scaled: %.2f", result[1].Close)
	}

	if !almostEqual(result[3].Close, 230.0, 0.01) {
		t.Errorf("High volume candle should be scaled: %.2f", result[3].Close)
	}
}

func TestConditionBuilders(t *testing.T) {
	candle := &market.Candle{
		Open:   100.0,
		High:   110.0,
		Low:    95.0,
		Close:  105.0,
		Volume: 1500.0,
	}

	tests := []struct {
		name      string
		condition Condition
		want      bool
	}{
		{"VolumeAbove", VolumeAbove(1000.0), true},
		{"VolumeBelow", VolumeBelow(2000.0), true},
		{"PriceAbove", PriceAbove(100.0), true},
		{"PriceBelow", PriceBelow(110.0), true},
		{"PriceInRange", PriceInRange(100.0, 110.0), true},
		{"BullishCandle", BullishCandle(), true},
		{"BearishCandle", BearishCandle(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.condition(candle)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestAndCondition(t *testing.T) {
	candle := &market.Candle{Close: 105.0, Volume: 2000.0}

	cond := AndCondition(
		PriceAbove(100.0),
		VolumeAbove(1500.0),
	)

	if !cond(candle) {
		t.Error("AND condition should be true")
	}

	cond2 := AndCondition(
		PriceAbove(100.0),
		VolumeAbove(3000.0),
	)

	if cond2(candle) {
		t.Error("AND condition should be false")
	}
}

func TestOrCondition(t *testing.T) {
	candle := &market.Candle{Close: 105.0, Volume: 1000.0}

	cond := OrCondition(
		PriceAbove(110.0),
		VolumeAbove(900.0),
	)

	if !cond(candle) {
		t.Error("OR condition should be true")
	}
}

func TestSwitchTransform(t *testing.T) {
	candles := []*market.Candle{
		{Close: 100.0, Volume: 1000.0}, // Low volume
		{Close: 105.0, Volume: 3000.0}, // High volume
		{Close: 110.0, Volume: 1500.0}, // Medium
	}

	st := NewSwitchTransform(NewScaleTransform(1.0, "close"))
	
	// High volume -> scale by 2
	st.AddCase(VolumeAbove(2500.0), NewScaleTransform(2.0, "close"))
	
	// Low volume -> scale by 0.5
	st.AddCase(VolumeBelow(1200.0), NewScaleTransform(0.5, "close"))

	result, err := st.Apply(candles)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Check results
	if !almostEqual(result[0].Close, 50.0, 0.01) {
		t.Errorf("Low volume: expected 50, got %.2f", result[0].Close)
	}

	if !almostEqual(result[1].Close, 210.0, 0.01) {
		t.Errorf("High volume: expected 210, got %.2f", result[1].Close)
	}

	if !almostEqual(result[2].Close, 110.0, 0.01) {
		t.Errorf("Default: expected 110, got %.2f", result[2].Close)
	}
}

func TestFilterTransform(t *testing.T) {
	candles := []*market.Candle{
		{Close: 100.0, Volume: 1000.0},
		{Close: 105.0, Volume: 2000.0},
		{Close: 110.0, Volume: 1500.0},
		{Close: 115.0, Volume: 3000.0},
	}

	// Filter high volume candles
	ft := NewFilterTransform(VolumeAbove(1800.0))

	result, err := ft.Apply(candles)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 filtered candles, got %d", len(result))
	}

	if result[0].Volume <= 1800.0 {
		t.Error("Filtered candle has low volume")
	}
}

// Parallel tests

func TestParallelMultiSymbolFeed(t *testing.T) {
	pmsf := NewParallelMultiSymbolFeed(2)

	// Add multiple symbols
	for _, symbol := range []string{"BTC", "ETH", "SOL"} {
		csvPath := createTestCSV(t, 20)
		config := csv.DefaultCSVConfig(market.Symbol(symbol), "1h")
		feed, _ := csv.NewCSVDataFeed(csvPath, config)
		chain := NewTransformChain(NewScaleTransform(2.0, "close"))
		
		err := pmsf.AddSymbol(symbol, feed, chain, 10)
		if err != nil {
			t.Fatalf("AddSymbol %s failed: %v", symbol, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := pmsf.ReadAllParallel(ctx)
	if err != nil {
		t.Fatalf("ReadAllParallel failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 symbols, got %d", len(results))
	}

	for symbol, candles := range results {
		if len(candles) != 20 {
			t.Errorf("%s: expected 20 candles, got %d", symbol, len(candles))
		}
	}

	pmsf.Close()
}
