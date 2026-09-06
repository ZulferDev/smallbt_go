package resample

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// TestIntegration_ForwardFillCorrectness verifies forward-fill with real patterns
func TestIntegration_ForwardFillCorrectness(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	btc := []*market.Candle{
		{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: t0.Add(1*time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		{Timestamp: t0.Add(2*time.Minute), Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
	}
	eth := []*market.Candle{
		{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
		// Missing t0+1min - should forward-fill from t0
		{Timestamp: t0.Add(2*time.Minute), Open: 55, High: 60, Low: 50, Close: 58, Volume: 2500},
	}
	
	aligner := NewDefaultAligner()
	result, err := aligner.Align(map[string][]*market.Candle{
		"BTC": btc,
		"ETH": eth,
	})
	
	if err != nil {
		t.Fatalf("Align failed: %v", err)
	}
	
	btcAligned := result["BTC"]
	ethAligned := result["ETH"]
	
	// Requirement 1: Length must match
	if len(btcAligned) != len(ethAligned) {
		t.Errorf("Length mismatch - BTC:%d ETH:%d", len(btcAligned), len(ethAligned))
	}
	if len(btcAligned) != 3 {
		t.Errorf("Expected 3 aligned candles, got %d", len(btcAligned))
	}
	
	// Requirement 2: Timestamps must be synchronized
	for i := range btcAligned {
		if !btcAligned[i].Timestamp.Equal(ethAligned[i].Timestamp) {
			t.Errorf("Timestamp mismatch at index %d: BTC=%v ETH=%v", 
				i, btcAligned[i].Timestamp, ethAligned[i].Timestamp)
		}
	}
	
	// Requirement 3: Forward-fill values must match last known
	eth1 := ethAligned[1]
	if eth1.Open != 50 {
		t.Errorf("Forward-fill Open incorrect: got %f, want 50", eth1.Open)
	}
	if eth1.Close != 52 {
		t.Errorf("Forward-fill Close incorrect: got %f, want 52", eth1.Close)
	}
	if eth1.High != 55 {
		t.Errorf("Forward-fill High incorrect: got %f, want 55", eth1.High)
	}
	if eth1.Low != 45 {
		t.Errorf("Forward-fill Low incorrect: got %f, want 45", eth1.Low)
	}
	
	// Requirement 4: Forward-fill timestamp must be updated
	if !eth1.Timestamp.Equal(t0.Add(1*time.Minute)) {
		t.Errorf("Forward-fill timestamp incorrect: got %v, want %v", 
			eth1.Timestamp, t0.Add(1*time.Minute))
	}
}

// TestIntegration_DropStrategyCompleteness verifies drop strategy removes incomplete rows
func TestIntegration_DropStrategyCompleteness(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	btc := []*market.Candle{
		{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: t0.Add(1*time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		{Timestamp: t0.Add(2*time.Minute), Open: 110, High: 120, Low: 105, Close: 115, Volume: 1200},
	}
	eth := []*market.Candle{
		{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
		// Missing t0+1min
		{Timestamp: t0.Add(2*time.Minute), Open: 55, High: 60, Low: 50, Close: 58, Volume: 2500},
	}
	
	aligner := NewDefaultAligner()
	aligner.FillStrategy = FillStrategyDrop
	
	result, err := aligner.Align(map[string][]*market.Candle{
		"BTC": btc,
		"ETH": eth,
	})
	
	if err != nil {
		t.Fatalf("Align failed: %v", err)
	}
	
	btcDropped := result["BTC"]
	ethDropped := result["ETH"]
	
	// Requirement 1: Only complete rows remain
	if len(btcDropped) != 2 {
		t.Errorf("BTC: expected 2 candles (complete rows), got %d", len(btcDropped))
	}
	if len(ethDropped) != 2 {
		t.Errorf("ETH: expected 2 candles (complete rows), got %d", len(ethDropped))
	}
	
	// Requirement 2: Only t0 and t2 should remain (t1 dropped)
	if !btcDropped[0].Timestamp.Equal(t0) {
		t.Errorf("First candle should be t0, got %v", btcDropped[0].Timestamp)
	}
	if !btcDropped[1].Timestamp.Equal(t0.Add(2*time.Minute)) {
		t.Errorf("Second candle should be t2, got %v", btcDropped[1].Timestamp)
	}
}

// TestIntegration_ReferenceBasedAlignment verifies reference-driven timeline
func TestIntegration_ReferenceBasedAlignment(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	btcRef := []*market.Candle{
		{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Timestamp: t0.Add(1*time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
	}
	ethLonger := []*market.Candle{
		{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
		{Timestamp: t0.Add(1*time.Minute), Open: 52, High: 58, Low: 48, Close: 55, Volume: 2500},
		{Timestamp: t0.Add(2*time.Minute), Open: 55, High: 60, Low: 50, Close: 58, Volume: 2700},
	}
	
	aligner := NewDefaultAligner()
	result, err := aligner.AlignToReference("BTC", map[string][]*market.Candle{
		"BTC": btcRef,
		"ETH": ethLonger,
	})
	
	if err != nil {
		t.Fatalf("AlignToReference failed: %v", err)
	}
	
	btcResult := result["BTC"]
	ethResult := result["ETH"]
	
	// Requirement 1: Reference unchanged
	if len(btcResult) != 2 {
		t.Errorf("BTC (reference): expected 2 candles, got %d", len(btcResult))
	}
	
	// Requirement 2: Other symbols match reference length
	if len(ethResult) != 2 {
		t.Errorf("ETH: expected 2 candles (matched to BTC), got %d", len(ethResult))
	}
	
	// Requirement 3: Extra timestamps excluded
	for _, candle := range ethResult {
		if candle.Timestamp.Equal(t0.Add(2*time.Minute)) {
			t.Error("ETH t2 should be excluded (not in reference)")
		}
	}
	
	// Requirement 4: All returned candles match reference timestamps
	for i := range ethResult {
		if !ethResult[i].Timestamp.Equal(btcResult[i].Timestamp) {
			t.Errorf("Timestamp mismatch at %d: BTC=%v ETH=%v",
				i, btcResult[i].Timestamp, ethResult[i].Timestamp)
		}
	}
}

// TestIntegration_PublicAPIContracts verifies public interface contracts
func TestIntegration_PublicAPIContracts(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// Contract 1: Empty input returns empty output (not error)
	aligner := NewDefaultAligner()
	result, err := aligner.Align(map[string][]*market.Candle{})
	if err != nil {
		t.Errorf("Empty input should not error, got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Empty input should return empty result, got %d symbols", len(result))
	}
	
	// Contract 2: Single symbol returns unchanged (except copy)
	single := []*market.Candle{
		{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}
	result2, err := aligner.Align(map[string][]*market.Candle{"BTC": single})
	if err != nil {
		t.Errorf("Single symbol should not error, got: %v", err)
	}
	if len(result2["BTC"]) != 1 {
		t.Errorf("Single symbol should return 1 candle, got %d", len(result2["BTC"]))
	}
	
	// Contract 3: FillStrategyNone errors on gaps
	aligner3 := NewDefaultAligner()
	aligner3.FillStrategy = FillStrategyNone
	
	gappedData := map[string][]*market.Candle{
		"BTC": {
			{Timestamp: t0, Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
			{Timestamp: t0.Add(1*time.Minute), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1500},
		},
		"ETH": {
			{Timestamp: t0, Open: 50, High: 55, Low: 45, Close: 52, Volume: 2000},
			// Missing t0+1min - should error with FillStrategyNone
		},
	}
	
	_, err = aligner3.Align(gappedData)
	if err == nil {
		t.Error("FillStrategyNone should error on missing data")
	}
}
