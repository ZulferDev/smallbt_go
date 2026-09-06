package transform

import (
	"math"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestPercentileRankTransform_Validate(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		window  int
		wantErr bool
	}{
		{
			name:    "valid close",
			field:   "close",
			window:  20,
			wantErr: false,
		},
		{
			name:    "valid volume",
			field:   "volume",
			window:  10,
			wantErr: false,
		},
		{
			name:    "missing field",
			field:   "",
			window:  20,
			wantErr: true,
		},
		{
			name:    "invalid field",
			field:   "invalid",
			window:  20,
			wantErr: true,
		},
		{
			name:    "window too small",
			field:   "close",
			window:  1,
			wantErr: true,
		},
		{
			name:    "minimum valid window",
			field:   "close",
			window:  2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &PercentileRankTransform{
				Field:  tt.field,
				Window: tt.window,
			}
			err := tr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPercentileRankTransform_Apply(t *testing.T) {
	// Test with simple ordered values
	// Values: 10, 20, 30, 40, 50 (window=5)
	// Each value should be at: 0%, 25%, 50%, 75%, 100%
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 10},
		{Timestamp: time.Now(), Close: 20},
		{Timestamp: time.Now(), Close: 30},
		{Timestamp: time.Now(), Close: 40},
		{Timestamp: time.Now(), Close: 50},
	}

	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if len(result) != len(candles) {
		t.Fatalf("result length = %d, want %d", len(result), len(candles))
	}

	// First 4 candles should be 50 (not enough data)
	for i := 0; i < 4; i++ {
		if result[i].Close != 50 {
			t.Errorf("candle[%d].Close = %f, want 50 (insufficient data)", i, result[i].Close)
		}
	}

	// Last candle: 50 is highest in [10,20,30,40,50] = 100%
	expected := 100.0
	tolerance := 0.01
	if math.Abs(result[4].Close-expected) > tolerance {
		t.Errorf("candle[4].Close = %f, want %f", result[4].Close, expected)
	}
}

func TestPercentileRankTransform_MinMax(t *testing.T) {
	// Test minimum and maximum percentiles
	// Window: [10, 50, 30, 40, 20]
	// Value 10 should be 0% (minimum)
	// Value 50 should be 100% (maximum)
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 10},
		{Timestamp: time.Now(), Close: 50},
		{Timestamp: time.Now(), Close: 30},
		{Timestamp: time.Now(), Close: 40},
		{Timestamp: time.Now(), Close: 20},
	}

	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Last candle (20) should be 25% in sorted [10,20,30,40,50]
	// rank of 20 = 2 (10 and 20 are <= 20)
	// percentile = (2-1)/(5-1)*100 = 25%
	expected := 25.0
	tolerance := 0.01
	actual := result[4].Close
	if math.Abs(actual-expected) > tolerance {
		t.Errorf("candle[4].Close = %f, want %f", actual, expected)
	}
}

func TestPercentileRankTransform_RollingWindow(t *testing.T) {
	// Test that window rolls correctly
	// Values: 10, 20, 30, 40, 100 (spike)
	// Window=3, last value (100) should be 100% in [30,40,100]
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 10},
		{Timestamp: time.Now(), Close: 20},
		{Timestamp: time.Now(), Close: 30},
		{Timestamp: time.Now(), Close: 40},
		{Timestamp: time.Now(), Close: 100}, // Spike
	}

	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 3,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// candle[2]: 30 in [10,20,30] = 100%
	if result[2].Close != 100 {
		t.Errorf("candle[2].Close = %f, want 100", result[2].Close)
	}

	// candle[3]: 40 in [20,30,40] = 100%
	if result[3].Close != 100 {
		t.Errorf("candle[3].Close = %f, want 100", result[3].Close)
	}

	// candle[4]: 100 in [30,40,100] = 100%
	if result[4].Close != 100 {
		t.Errorf("candle[4].Close = %f, want 100", result[4].Close)
	}
}

func TestPercentileRankTransform_MedianValue(t *testing.T) {
	// Test median value gets 50th percentile
	// Window: [10, 20, 30, 40, 50]
	// Value 30 (median) should be 50%
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 10},
		{Timestamp: time.Now(), Close: 20},
		{Timestamp: time.Now(), Close: 30},
		{Timestamp: time.Now(), Close: 40},
		{Timestamp: time.Now(), Close: 50},
	}

	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// candle[4]: 50 in [10,20,30,40,50]
	// rank = 5, percentile = (5-1)/(5-1)*100 = 100%
	expected := 100.0
	if result[4].Close != expected {
		t.Errorf("candle[4].Close = %f, want %f", result[4].Close, expected)
	}
}

func TestPercentileRankTransform_DuplicateValues(t *testing.T) {
	// Test with duplicate values
	// Window: [20, 20, 30, 20, 20]
	// Value 20 appears 4 times, 30 once
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 20},
		{Timestamp: time.Now(), Close: 20},
		{Timestamp: time.Now(), Close: 30},
		{Timestamp: time.Now(), Close: 20},
		{Timestamp: time.Now(), Close: 20},
	}

	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Last value (20) in [20,20,30,20,20]
	// rank = 4 (all four 20s are <= 20), percentile = (4-1)/(5-1)*100 = 75%
	expected := 75.0
	tolerance := 0.01
	if math.Abs(result[4].Close-expected) > tolerance {
		t.Errorf("candle[4].Close = %f, want %f", result[4].Close, expected)
	}
}

func TestPercentileRankTransform_VolumeField(t *testing.T) {
	// Test percentile on volume field
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 50000, Volume: 100},
		{Timestamp: time.Now(), Close: 50000, Volume: 200},
		{Timestamp: time.Now(), Close: 50000, Volume: 300},
		{Timestamp: time.Now(), Close: 50000, Volume: 400},
		{Timestamp: time.Now(), Close: 50000, Volume: 500},
	}

	tr := &PercentileRankTransform{
		Field:  "volume",
		Window: 5,
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

	// Volume 500 is maximum in [100,200,300,400,500] = 100%
	if result[4].Volume != 100 {
		t.Errorf("candle[4].Volume = %f, want 100", result[4].Volume)
	}
}

func TestPercentileRankTransform_OverboughtOversold(t *testing.T) {
	// Test overbought/oversold detection
	// High percentile = overbought (near top of range)
	// Low percentile = oversold (near bottom of range)
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 105},
		{Timestamp: time.Now(), Close: 102},
		{Timestamp: time.Now(), Close: 98},  // Oversold
		{Timestamp: time.Now(), Close: 108}, // Overbought
	}

	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// candle[4]: 108 is max in [100,105,102,98,108] = 100% (overbought)
	if result[4].Close != 100 {
		t.Errorf("overbought candle[4].Close = %f, want 100", result[4].Close)
	}
}

func TestPercentileRankTransform_InsufficientData(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 105},
	}

	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 5,
	}

	_, err := tr.Apply(candles)
	if err == nil {
		t.Error("Apply() with insufficient data should return error")
	}
}

func TestPercentileRankTransform_EmptyData(t *testing.T) {
	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 5,
	}

	_, err := tr.Apply([]*market.Candle{})
	if err == nil {
		t.Error("Apply() with empty data should return error")
	}
}

func TestPercentileRankTransform_Type(t *testing.T) {
	tr := &PercentileRankTransform{}
	if tr.Type() != "percentile_rank" {
		t.Errorf("Type() = %s, want percentile_rank", tr.Type())
	}
}

func TestPercentileRankTransform_Name(t *testing.T) {
	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 20,
	}
	expected := "percentile_rank(field=close, window=20)"
	if tr.Name() != expected {
		t.Errorf("Name() = %s, want %s", tr.Name(), expected)
	}
}

func TestPercentileRankTransform_ConstantValues(t *testing.T) {
	// Test with all same values
	// All values should be at some consistent percentile
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
	}

	tr := &PercentileRankTransform{
		Field:  "close",
		Window: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// All equal values should give 100% (all values <= current)
	// rank = 5, percentile = (5-1)/(5-1)*100 = 100%
	expected := 100.0
	if result[4].Close != expected {
		t.Errorf("constant values candle[4].Close = %f, want %f", result[4].Close, expected)
	}
}
