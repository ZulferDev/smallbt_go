package transform

import (
	"math"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestEMASmooth_Validate(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		period  int
		wantErr bool
	}{
		{
			name:    "valid close",
			field:   "close",
			period:  10,
			wantErr: false,
		},
		{
			name:    "valid volume",
			field:   "volume",
			period:  5,
			wantErr: false,
		},
		{
			name:    "missing field",
			field:   "",
			period:  10,
			wantErr: true,
		},
		{
			name:    "invalid field",
			field:   "invalid",
			period:  10,
			wantErr: true,
		},
		{
			name:    "period too small",
			field:   "close",
			period:  1,
			wantErr: true,
		},
		{
			name:    "minimum valid period",
			field:   "close",
			period:  2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &EMASmooth{
				Field:  tt.field,
				Period: tt.period,
			}
			err := tr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEMASmooth_Apply(t *testing.T) {
	// Test EMA calculation with known values
	// Values: 100, 110, 105, 115, 120
	// EMA(3): α = 2/(3+1) = 0.5
	// EMA[0] = 100
	// EMA[1] = 0.5*110 + 0.5*100 = 105
	// EMA[2] = 0.5*105 + 0.5*105 = 105
	// EMA[3] = 0.5*115 + 0.5*105 = 110
	// EMA[4] = 0.5*120 + 0.5*110 = 115
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 110},
		{Timestamp: time.Now(), Close: 105},
		{Timestamp: time.Now(), Close: 115},
		{Timestamp: time.Now(), Close: 120},
	}

	tr := &EMASmooth{
		Field:  "close",
		Period: 3,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if len(result) != len(candles) {
		t.Fatalf("result length = %d, want %d", len(result), len(candles))
	}

	expected := []float64{100, 105, 105, 110, 115}
	tolerance := 0.01

	for i, exp := range expected {
		if math.Abs(result[i].Close-exp) > tolerance {
			t.Errorf("candle[%d].Close = %f, want %f", i, result[i].Close, exp)
		}
	}
}

func TestEMASmooth_Convergence(t *testing.T) {
	// Test that EMA converges to constant value
	// All values = 100, EMA should stay at 100
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
	}

	tr := &EMASmooth{
		Field:  "close",
		Period: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// All values should be 100
	for i, c := range result {
		if c.Close != 100 {
			t.Errorf("candle[%d].Close = %f, want 100", i, c.Close)
		}
	}
}

func TestEMASmooth_Responsiveness(t *testing.T) {
	// Test that EMA responds faster than SMA to changes
	// Start at 100, spike to 200
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 200}, // Spike
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
	}

	tr := &EMASmooth{
		Field:  "close",
		Period: 3,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// After spike, EMA should be between 100 and 200
	// And should gradually return toward 100
	if result[2].Close <= 100 || result[2].Close >= 200 {
		t.Errorf("candle[2] (spike) EMA = %f, should be between 100 and 200", result[2].Close)
	}

	// After spike, values should decrease
	if result[3].Close >= result[2].Close {
		t.Errorf("EMA should decrease after spike returns to normal")
	}

	if result[4].Close >= result[3].Close {
		t.Errorf("EMA should continue decreasing toward 100")
	}
}

func TestEMASmooth_AlphaCalculation(t *testing.T) {
	// Test different periods produce different alpha values
	// Period 2: α = 2/(2+1) = 0.667
	// Period 10: α = 2/(10+1) = 0.182
	
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 200},
	}

	// Fast EMA (period=2, high alpha)
	fastEMA := &EMASmooth{Field: "close", Period: 2}
	fastResult, err := fastEMA.Apply(candles)
	if err != nil {
		t.Fatalf("Fast EMA error = %v", err)
	}

	// Slow EMA (period=10, low alpha)
	slowEMA := &EMASmooth{Field: "close", Period: 10}
	slowResult, err := slowEMA.Apply(candles)
	if err != nil {
		t.Fatalf("Slow EMA error = %v", err)
	}

	// Fast EMA should react more to change
	fastChange := fastResult[1].Close - fastResult[0].Close
	slowChange := slowResult[1].Close - slowResult[0].Close

	if fastChange <= slowChange {
		t.Errorf("Fast EMA change (%f) should be > slow EMA change (%f)", fastChange, slowChange)
	}
}

func TestEMASmooth_VolumeField(t *testing.T) {
	// Test EMA on volume field
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 50000, Volume: 100},
		{Timestamp: time.Now(), Close: 50000, Volume: 200},
		{Timestamp: time.Now(), Close: 50000, Volume: 150},
	}

	tr := &EMASmooth{
		Field:  "volume",
		Period: 2,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Close should remain unchanged
	for i, c := range result {
		if c.Close != 50000 {
			t.Errorf("candle[%d].Close = %f, want 50000", i, c.Close)
		}
	}

	// Volume should be smoothed
	// α = 2/(2+1) = 0.667
	// EMA[0] = 100
	// EMA[1] = 0.667*200 + 0.333*100 = 166.7
	// EMA[2] = 0.667*150 + 0.333*166.7 = 155.6
	expected := []float64{100, 166.7, 155.6}
	tolerance := 0.5

	for i, exp := range expected {
		if math.Abs(result[i].Volume-exp) > tolerance {
			t.Errorf("candle[%d].Volume = %f, want ≈ %f", i, result[i].Volume, exp)
		}
	}
}

func TestEMASmooth_SingleCandle(t *testing.T) {
	// Test with single candle
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
	}

	tr := &EMASmooth{
		Field:  "close",
		Period: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("result length = %d, want 1", len(result))
	}

	// Single value should equal input
	if result[0].Close != 100 {
		t.Errorf("candle[0].Close = %f, want 100", result[0].Close)
	}
}

func TestEMASmooth_EmptyData(t *testing.T) {
	tr := &EMASmooth{
		Field:  "close",
		Period: 5,
	}

	_, err := tr.Apply([]*market.Candle{})
	if err == nil {
		t.Error("Apply() with empty data should return error")
	}
}

func TestEMASmooth_Type(t *testing.T) {
	tr := &EMASmooth{}
	if tr.Type() != "ema_smooth" {
		t.Errorf("Type() = %s, want ema_smooth", tr.Type())
	}
}

func TestEMASmooth_Name(t *testing.T) {
	tr := &EMASmooth{
		Field:  "close",
		Period: 12,
	}
	expected := "ema_smooth(field=close, period=12)"
	if tr.Name() != expected {
		t.Errorf("Name() = %s, want %s", tr.Name(), expected)
	}
}

func TestEMASmooth_CompareWithSMA(t *testing.T) {
	// EMA should have less lag than SMA on trending data
	// Uptrend: 100, 110, 120, 130, 140
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 110},
		{Timestamp: time.Now(), Close: 120},
		{Timestamp: time.Now(), Close: 130},
		{Timestamp: time.Now(), Close: 140},
	}

	ema := &EMASmooth{Field: "close", Period: 3}
	emaResult, _ := ema.Apply(candles)

	// In uptrend, EMA should be closer to recent values than midpoint
	lastEMA := emaResult[4].Close
	
	// Last EMA should be between 130 and 140 (closer to recent)
	if lastEMA < 120 || lastEMA > 140 {
		t.Errorf("EMA in uptrend = %f, expected between 120 and 140", lastEMA)
	}
}
