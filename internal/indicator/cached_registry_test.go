package indicator

import (
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestCachedRegistry_PrefersCachedVersion(t *testing.T) {
	reg := BuiltinCachedRegistry()

	// ATR should use cached version
	config := Config{
		Type:   "atr",
		Period: 14,
	}

	ind, err := reg.CreateCached(config)
	if err != nil {
		t.Fatalf("CreateCached error: %v", err)
	}

	// Should be CachedATR, not wrapped
	if _, ok := ind.(*CachedATR); !ok {
		t.Errorf("expected *CachedATR, got %T", ind)
	}
}

func TestCachedRegistry_FallbackToStateless(t *testing.T) {
	reg := BuiltinCachedRegistry()

	// SMA doesn't have cached version yet, should be wrapped
	config := Config{
		Type:   "sma",
		Period: 20,
		Source: "close",
	}

	ind, err := reg.CreateCached(config)
	if err != nil {
		t.Fatalf("CreateCached error: %v", err)
	}

	// Should be statelessWrapper
	if _, ok := ind.(*statelessWrapper); !ok {
		t.Errorf("expected *statelessWrapper, got %T", ind)
	}

	// Verify it works
	candles := generateTestCandles(30)
	var prevCandle *market.Candle

	for i, candle := range candles {
		val, err := ind.Update(candle, prevCandle)
		if err != nil {
			t.Fatalf("Update error at bar %d: %v", i, err)
		}

		if i >= 20 && !val.Valid {
			t.Errorf("bar %d: expected valid value", i)
		}

		prevCandle = &candle
	}
}

func TestCachedRegistry_UnknownIndicator(t *testing.T) {
	reg := BuiltinCachedRegistry()

	config := Config{
		Type:   "unknown",
		Period: 14,
	}

	_, err := reg.CreateCached(config)
	if err == nil {
		t.Error("expected error for unknown indicator type")
	}
}

func TestCachedRegistry_HasCached(t *testing.T) {
	reg := BuiltinCachedRegistry()

	if !reg.HasCached("atr") {
		t.Error("expected atr to have cached version")
	}

	if reg.HasCached("sma") {
		t.Error("expected sma to not have cached version yet")
	}
}

func TestStatelessWrapper_MatchesOriginal(t *testing.T) {
	// Create stateless SMA
	statelessConfig := Config{
		Type:   "sma",
		Period: 10,
		Source: "close",
	}
	stateless, err := SMAFactory(statelessConfig)
	if err != nil {
		t.Fatalf("SMAFactory error: %v", err)
	}

	// Create wrapped version
	wrapper := &statelessWrapper{
		indicator: stateless,
		period:    10,
	}

	candles := generateTestCandles(30)
	var prevCandle *market.Candle

	// Update wrapper
	var wrapperValues []Value
	for _, candle := range candles {
		val, err := wrapper.Update(candle, prevCandle)
		if err != nil {
			t.Fatalf("wrapper.Update error: %v", err)
		}
		wrapperValues = append(wrapperValues, val)
		prevCandle = &candle
	}

	// Calculate stateless
	var statelessValues []Value
	for i := range candles {
		ctx := &Context{
			Current:  candles[i],
			Candles:  candles[:i+1],
			BarIndex: i,
		}
		val, err := stateless.Calculate(ctx)
		if err != nil {
			t.Fatalf("stateless.Calculate error: %v", err)
		}
		statelessValues = append(statelessValues, val)
	}

	// Compare
	for i := range candles {
		if wrapperValues[i].Valid != statelessValues[i].Valid {
			t.Errorf("bar %d: validity mismatch", i)
		}
		if wrapperValues[i].Valid && statelessValues[i].Valid {
			diff := wrapperValues[i].Value - statelessValues[i].Value
			if diff > 0.0001 || diff < -0.0001 {
				t.Errorf("bar %d: value mismatch: wrapper=%.6f stateless=%.6f",
					i, wrapperValues[i].Value, statelessValues[i].Value)
			}
		}
	}
}
