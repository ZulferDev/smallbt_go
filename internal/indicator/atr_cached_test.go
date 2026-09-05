package indicator

import (
	"math"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// TestCachedATR_MatchesStateless verifies that CachedATR produces identical results
// to the stateless ATR implementation (golden test).
func TestCachedATR_MatchesStateless(t *testing.T) {
	// Generate test candles
	candles := generateTestCandles(100)
	period := 14

	// Calculate using stateless ATR
	statelessATR := &ATR{name: "atr", period: period}
	statelessResults := make([]Value, len(candles))

	for i := range candles {
		ctx := &Context{
			Current:  candles[i],
			Candles:  candles[:i+1],
			BarIndex: i,
		}
		val, err := statelessATR.Calculate(ctx)
		if err != nil {
			t.Fatalf("stateless ATR error at bar %d: %v", i, err)
		}
		statelessResults[i] = val
	}

	// Calculate using cached ATR
	cachedATR, err := NewCachedATR(period)
	if err != nil {
		t.Fatalf("NewCachedATR error: %v", err)
	}

	cachedResults := make([]Value, len(candles))
	var prevCandle *market.Candle

	for i, candle := range candles {
		val, err := cachedATR.Update(candle, prevCandle)
		if err != nil {
			t.Fatalf("cached ATR error at bar %d: %v", i, err)
		}
		cachedResults[i] = val
		prevCandle = &candle
	}

	// Compare results
	for i := range candles {
		stateless := statelessResults[i]
		cached := cachedResults[i]

		if stateless.Valid != cached.Valid {
			t.Errorf("bar %d: validity mismatch: stateless=%v cached=%v",
				i, stateless.Valid, cached.Valid)
		}

		if stateless.Valid && cached.Valid {
			diff := math.Abs(stateless.Value - cached.Value)
			if diff > 0.0001 { // Allow small floating point differences
				t.Errorf("bar %d: value mismatch: stateless=%.6f cached=%.6f diff=%.6f",
					i, stateless.Value, cached.Value, diff)
			}
		}
	}
}

// TestCachedATR_WarmupPeriod verifies warmup behavior.
func TestCachedATR_WarmupPeriod(t *testing.T) {
	period := 14
	cachedATR, err := NewCachedATR(period)
	if err != nil {
		t.Fatalf("NewCachedATR error: %v", err)
	}

	if cachedATR.WarmupPeriod() != period+1 {
		t.Errorf("WarmupPeriod: expected %d, got %d", period+1, cachedATR.WarmupPeriod())
	}

	candles := generateTestCandles(20)
	var prevCandle *market.Candle

	for i, candle := range candles {
		val, err := cachedATR.Update(candle, prevCandle)
		if err != nil {
			t.Fatalf("Update error at bar %d: %v", i, err)
		}

		// Should be invalid until we have period+1 bars
		if i < period {
			if val.Valid {
				t.Errorf("bar %d: expected invalid, got valid", i)
			}
			if cachedATR.IsWarm() {
				t.Errorf("bar %d: expected not warm", i)
			}
		} else {
			if !val.Valid {
				t.Errorf("bar %d: expected valid, got invalid", i)
			}
			if !cachedATR.IsWarm() {
				t.Errorf("bar %d: expected warm", i)
			}
		}

		prevCandle = &candle
	}
}

// TestCachedATR_Reset verifies that Reset clears state.
func TestCachedATR_Reset(t *testing.T) {
	cachedATR, err := NewCachedATR(14)
	if err != nil {
		t.Fatalf("NewCachedATR error: %v", err)
	}

	candles := generateTestCandles(20)
	var prevCandle *market.Candle

	// Update to warm state
	for _, candle := range candles {
		cachedATR.Update(candle, prevCandle)
		prevCandle = &candle
	}

	if !cachedATR.IsWarm() {
		t.Fatal("expected warm after updates")
	}

	// Reset
	cachedATR.Reset()

	if cachedATR.IsWarm() {
		t.Error("expected not warm after reset")
	}
	if cachedATR.barsSeen != 0 {
		t.Errorf("barsSeen after reset: expected 0, got %d", cachedATR.barsSeen)
	}
	if cachedATR.currentATR != 0 {
		t.Errorf("currentATR after reset: expected 0, got %.6f", cachedATR.currentATR)
	}

	// Verify we can update again after reset
	prevCandle = nil
	for i, candle := range candles[:5] {
		_, err := cachedATR.Update(candle, prevCandle)
		if err != nil {
			t.Fatalf("Update after reset error at bar %d: %v", i, err)
		}
		prevCandle = &candle
	}
}

// TestCachedATR_Determinism verifies same input produces same output.
func TestCachedATR_Determinism(t *testing.T) {
	candles := generateTestCandles(50)
	period := 14

	// Run 1
	atr1, _ := NewCachedATR(period)
	results1 := make([]Value, len(candles))
	var prev1 *market.Candle
	for i, candle := range candles {
		results1[i], _ = atr1.Update(candle, prev1)
		prev1 = &candle
	}

	// Run 2 (same candles)
	atr2, _ := NewCachedATR(period)
	results2 := make([]Value, len(candles))
	var prev2 *market.Candle
	for i, candle := range candles {
		results2[i], _ = atr2.Update(candle, prev2)
		prev2 = &candle
	}

	// Compare
	for i := range results1 {
		if results1[i].Valid != results2[i].Valid {
			t.Errorf("bar %d: validity differs between runs", i)
		}
		if results1[i].Valid && results2[i].Valid {
			if results1[i].Value != results2[i].Value {
				t.Errorf("bar %d: value differs: run1=%.6f run2=%.6f",
					i, results1[i].Value, results2[i].Value)
			}
		}
	}
}

// TestStateManager verifies StateManager functionality.
func TestStateManager(t *testing.T) {
	sm := NewStateManager()

	atr1, _ := NewCachedATR(14)
	atr2, _ := NewCachedATR(20)

	sm.Register("atr14", atr1)
	sm.Register("atr20", atr2)

	if sm.MaxWarmupPeriod() != 21 { // atr20 needs 21 bars
		t.Errorf("MaxWarmupPeriod: expected 21, got %d", sm.MaxWarmupPeriod())
	}

	candles := generateTestCandles(30)
	var prevCandle *market.Candle

	for i, candle := range candles {
		err := sm.Update(candle, prevCandle)
		if err != nil {
			t.Fatalf("StateManager.Update error at bar %d: %v", i, err)
		}

		val14 := sm.GetValue("atr14")
		val20 := sm.GetValue("atr20")

		// Check warmup
		if i < 14 {
			if val14.Valid {
				t.Errorf("bar %d: atr14 should not be valid yet", i)
			}
		} else {
			if !val14.Valid {
				t.Errorf("bar %d: atr14 should be valid", i)
			}
		}

		if i < 20 {
			if val20.Valid {
				t.Errorf("bar %d: atr20 should not be valid yet", i)
			}
		} else {
			if !val20.Valid {
				t.Errorf("bar %d: atr20 should be valid", i)
			}
		}

		prevCandle = &candle
	}

	if !sm.AllWarm() {
		t.Error("expected all indicators warm after 30 bars")
	}

	// Test Reset
	sm.Reset()
	if sm.AllWarm() {
		t.Error("expected not all warm after reset")
	}
}

// generateTestCandles creates synthetic candles for testing.
func generateTestCandles(count int) []market.Candle {
	candles := make([]market.Candle, count)
	basePrice := 30000.0

	for i := range candles {
		// Simulate price movement
		volatility := 100.0 + float64(i)*2.0
		open := basePrice + float64(i)*10.0
		high := open + volatility
		low := open - volatility
		close := open + volatility*0.5

		candles[i] = market.Candle{
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: 1000.0,
		}

		basePrice = close
	}

	return candles
}
