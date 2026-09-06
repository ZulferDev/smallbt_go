package resample

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestNewDefaultAligner(t *testing.T) {
	a := NewDefaultAligner()
	if a == nil {
		t.Fatal("expected non-nil aligner")
	}
	if a.FillStrategy != FillStrategyForward {
		t.Errorf("expected forward-fill strategy, got %d", a.FillStrategy)
	}
}

func TestAligner_EmptyInput(t *testing.T) {
	a := NewDefaultAligner()
	result, err := a.Align(map[string][]*market.Candle{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d symbols", len(result))
	}
}

func TestAligner_SingleSymbol(t *testing.T) {
	a := NewDefaultAligner()
	
	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		},
	}

	result, err := a.Align(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(result))
	}

	btc := result["BTC"]
	if len(btc) != 2 {
		t.Errorf("expected 2 candles, got %d", len(btc))
	}
}

func TestAligner_TwoSymbols_PerfectAlignment(t *testing.T) {
	a := NewDefaultAligner()

	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		},
		"ETH": {
			{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			{Timestamp: t1, Open: 52, High: 58, Low: 48, Close: 55, Volume: 2500},
		},
	}

	result, err := a.Align(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(result))
	}

	btc := result["BTC"]
	eth := result["ETH"]

	if len(btc) != 2 {
		t.Errorf("BTC: expected 2 candles, got %d", len(btc))
	}
	if len(eth) != 2 {
		t.Errorf("ETH: expected 2 candles, got %d", len(eth))
	}

	// Verify timestamps match
	for i := range btc {
		if !btc[i].Timestamp.Equal(eth[i].Timestamp) {
			t.Errorf("timestamp mismatch at index %d: BTC=%v, ETH=%v", i, btc[i].Timestamp, eth[i].Timestamp)
		}
	}
}

func TestAligner_ForwardFill_OneMissing(t *testing.T) {
	a := NewDefaultAligner()
	a.FillStrategy = FillStrategyForward

	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
			{Timestamp: t2, Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
		},
		"ETH": {
			{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			// Missing t1 - should forward-fill from t0
			{Timestamp: t2, Open: 52, High: 58, Low: 48, Close: 55, Volume: 2500},
		},
	}

	result, err := a.Align(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	btc := result["BTC"]
	eth := result["ETH"]

	if len(btc) != 3 {
		t.Fatalf("BTC: expected 3 candles, got %d", len(btc))
	}
	if len(eth) != 3 {
		t.Fatalf("ETH: expected 3 candles, got %d", len(eth))
	}

	// Verify ETH at t1 is forward-filled from t0
	eth1 := eth[1]
	if !eth1.Timestamp.Equal(t1) {
		t.Errorf("ETH[1]: expected timestamp %v, got %v", t1, eth1.Timestamp)
	}
	if eth1.Open != 50 || eth1.Close != 52 {
		t.Errorf("ETH[1]: expected forward-filled values (O:50, C:52), got (O:%f, C:%f)", eth1.Open, eth1.Close)
	}
}

func TestAligner_ForwardFill_MultipleMissing(t *testing.T) {
	a := NewDefaultAligner()
	a.FillStrategy = FillStrategyForward

	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 1, 0, 3, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
			{Timestamp: t2, Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
			{Timestamp: t3, Open: 115, High: 125, Low: 110, Close: 120, Volume: 1300},
		},
		"ETH": {
			{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			// Missing t1, t2 - should forward-fill from t0
			{Timestamp: t3, Open: 52, High: 58, Low: 48, Close: 55, Volume: 2500},
		},
	}

	result, err := a.Align(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eth := result["ETH"]

	if len(eth) != 4 {
		t.Fatalf("ETH: expected 4 candles, got %d", len(eth))
	}

	// Verify ETH at t1 and t2 are forward-filled from t0
	for i, ts := range []time.Time{t1, t2} {
		candle := eth[i+1]
		if !candle.Timestamp.Equal(ts) {
			t.Errorf("ETH[%d]: expected timestamp %v, got %v", i+1, ts, candle.Timestamp)
		}
		if candle.Open != 50 || candle.Close != 52 {
			t.Errorf("ETH[%d]: expected forward-filled values", i+1)
		}
	}

	// Verify ETH at t3 is actual data
	eth3 := eth[3]
	if eth3.Open != 52 || eth3.Close != 55 {
		t.Errorf("ETH[3]: expected actual values (O:52, C:55), got (O:%f, C:%f)", eth3.Open, eth3.Close)
	}
}

func TestAligner_DropStrategy(t *testing.T) {
	a := NewDefaultAligner()
	a.FillStrategy = FillStrategyDrop

	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
			{Timestamp: t2, Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
		},
		"ETH": {
			{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			// Missing t1 - should drop this timestamp
			{Timestamp: t2, Open: 52, High: 58, Low: 48, Close: 55, Volume: 2500},
		},
	}

	result, err := a.Align(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	btc := result["BTC"]
	eth := result["ETH"]

	// Only t0 and t2 should remain (t1 dropped)
	if len(btc) != 2 {
		t.Errorf("BTC: expected 2 candles, got %d", len(btc))
	}
	if len(eth) != 2 {
		t.Errorf("ETH: expected 2 candles, got %d", len(eth))
	}

	// Verify timestamps
	if !btc[0].Timestamp.Equal(t0) || !btc[1].Timestamp.Equal(t2) {
		t.Errorf("BTC: unexpected timestamps")
	}
	if !eth[0].Timestamp.Equal(t0) || !eth[1].Timestamp.Equal(t2) {
		t.Errorf("ETH: unexpected timestamps")
	}
}

func TestAligner_NoneStrategy_Error(t *testing.T) {
	a := NewDefaultAligner()
	a.FillStrategy = FillStrategyNone

	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		},
		"ETH": {
			{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			// Missing t1 - should error with FillStrategyNone
		},
	}

	_, err := a.Align(input)
	if err == nil {
		t.Error("expected error with FillStrategyNone and missing data")
	}
}

func TestAligner_AlignToReference(t *testing.T) {
	a := NewDefaultAligner()

	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		},
		"ETH": {
			{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			{Timestamp: t1, Open: 52, High: 58, Low: 48, Close: 55, Volume: 2500},
			{Timestamp: t2, Open: 55, High: 60, Low: 50, Close: 58, Volume: 2700},
		},
	}

	result, err := a.AlignToReference("BTC", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	btc := result["BTC"]
	eth := result["ETH"]

	// Should have 2 timestamps (from BTC reference)
	if len(btc) != 2 {
		t.Errorf("BTC: expected 2 candles, got %d", len(btc))
	}
	if len(eth) != 2 {
		t.Errorf("ETH: expected 2 candles, got %d", len(eth))
	}

	// ETH at t2 should not be included (not in BTC)
	for _, candle := range eth {
		if candle.Timestamp.Equal(t2) {
			t.Error("ETH: t2 should not be included (not in reference)")
		}
	}
}

func TestAligner_AlignToReference_WithForwardFill(t *testing.T) {
	a := NewDefaultAligner()

	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
			{Timestamp: t2, Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
		},
		"ETH": {
			{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			// Missing t1, t2 - should forward-fill
		},
	}

	result, err := a.AlignToReference("BTC", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eth := result["ETH"]

	if len(eth) != 3 {
		t.Fatalf("ETH: expected 3 candles, got %d", len(eth))
	}

	// Verify forward-fill at t1 and t2
	for i := 1; i <= 2; i++ {
		candle := eth[i]
		if candle.Open != 50 || candle.Close != 52 {
			t.Errorf("ETH[%d]: expected forward-filled values", i)
		}
	}
}

func TestGetCommonTimeRange(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 1, 0, 3, 0, 0, time.UTC)
	t4 := time.Date(2024, 1, 1, 0, 4, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
			{Timestamp: t2, Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
			{Timestamp: t3, Open: 115, High: 125, Low: 110, Close: 120, Volume: 1300},
		},
		"ETH": {
			{Timestamp: t1, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000}, // Starts later
			{Timestamp: t2, Open: 52, High: 58, Low: 48, Close: 55, Volume: 2500},
			{Timestamp: t3, Open: 55, High: 60, Low: 50, Close: 58, Volume: 2700},
			{Timestamp: t4, Open: 58, High: 65, Low: 55, Close: 62, Volume: 2900}, // Ends later
		},
	}

	start, end, hasData := GetCommonTimeRange(input)

	if !hasData {
		t.Fatal("expected common time range")
	}

	// Common range: t1 to t3 (intersection)
	if !start.Equal(t1) {
		t.Errorf("expected start %v, got %v", t1, start)
	}
	if !end.Equal(t3) {
		t.Errorf("expected end %v, got %v", t3, end)
	}
}

func TestGetCommonTimeRange_NoOverlap(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 1, 0, 3, 0, 0, time.UTC)

	input := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		},
		"ETH": {
			{Timestamp: t2, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			{Timestamp: t3, Open: 52, High: 58, Low: 48, Close: 55, Volume: 2500},
		},
	}

	_, _, hasData := GetCommonTimeRange(input)

	if hasData {
		t.Error("expected no common time range (no overlap)")
	}
}

func TestFilterByTimeRange(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 1, 0, 3, 0, 0, time.UTC)
	t4 := time.Date(2024, 1, 1, 0, 4, 0, 0, time.UTC)

	candles := []*market.Candle{
		{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: t1, Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		{Timestamp: t2, Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
		{Timestamp: t3, Open: 115, High: 125, Low: 110, Close: 120, Volume: 1300},
		{Timestamp: t4, Open: 120, High: 130, Low: 115, Close: 125, Volume: 1400},
	}

	filtered := FilterByTimeRange(candles, t1, t3)

	if len(filtered) != 3 {
		t.Fatalf("expected 3 candles, got %d", len(filtered))
	}

	// Verify range: t1, t2, t3
	expectedTimes := []time.Time{t1, t2, t3}
	for i, expected := range expectedTimes {
		if !filtered[i].Timestamp.Equal(expected) {
			t.Errorf("candle[%d]: expected timestamp %v, got %v", i, expected, filtered[i].Timestamp)
		}
	}
}

func TestSortTimestamps(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 1, 0, 3, 0, 0, time.UTC)

	// Unsorted
	timestamps := []time.Time{t2, t0, t3, t1}

	sortTimestamps(timestamps)

	// Should be sorted: t0, t1, t2, t3
	expected := []time.Time{t0, t1, t2, t3}
	for i, exp := range expected {
		if !timestamps[i].Equal(exp) {
			t.Errorf("index %d: expected %v, got %v", i, exp, timestamps[i])
		}
	}
}
