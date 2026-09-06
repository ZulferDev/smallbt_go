package transform

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestClipTransform_Validate(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		min     float64
		max     float64
		wantErr bool
	}{
		{
			name:    "valid close",
			field:   "close",
			min:     0,
			max:     100,
			wantErr: false,
		},
		{
			name:    "valid volume",
			field:   "volume",
			min:     100,
			max:     10000,
			wantErr: false,
		},
		{
			name:    "missing field",
			field:   "",
			min:     0,
			max:     100,
			wantErr: true,
		},
		{
			name:    "invalid field",
			field:   "invalid",
			min:     0,
			max:     100,
			wantErr: true,
		},
		{
			name:    "min >= max",
			field:   "close",
			min:     100,
			max:     100,
			wantErr: true,
		},
		{
			name:    "min > max",
			field:   "close",
			min:     100,
			max:     50,
			wantErr: true,
		},
		{
			name:    "negative range valid",
			field:   "close",
			min:     -100,
			max:     100,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &ClipTransform{
				Field: tt.field,
				Min:   tt.min,
				Max:   tt.max,
			}
			err := tr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClipTransform_Apply(t *testing.T) {
	// Test clipping with range [50, 150]
	// Values: 10, 50, 100, 150, 200
	// Expected: 50, 50, 100, 150, 150
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 10},  // Below min
		{Timestamp: time.Now(), Close: 50},  // At min
		{Timestamp: time.Now(), Close: 100}, // In range
		{Timestamp: time.Now(), Close: 150}, // At max
		{Timestamp: time.Now(), Close: 200}, // Above max
	}

	tr := &ClipTransform{
		Field: "close",
		Min:   50,
		Max:   150,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if len(result) != len(candles) {
		t.Fatalf("result length = %d, want %d", len(result), len(candles))
	}

	expected := []float64{50, 50, 100, 150, 150}
	for i, exp := range expected {
		if result[i].Close != exp {
			t.Errorf("candle[%d].Close = %f, want %f", i, result[i].Close, exp)
		}
	}
}

func TestClipTransform_NoClipping(t *testing.T) {
	// Test when all values are within range
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 60},
		{Timestamp: time.Now(), Close: 80},
		{Timestamp: time.Now(), Close: 100},
	}

	tr := &ClipTransform{
		Field: "close",
		Min:   50,
		Max:   150,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// All values should remain unchanged
	for i, c := range result {
		if c.Close != candles[i].Close {
			t.Errorf("candle[%d].Close = %f, want %f (no clipping)", i, c.Close, candles[i].Close)
		}
	}
}

func TestClipTransform_OnlyLowerBound(t *testing.T) {
	// Test clipping only at lower bound
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 5},
		{Timestamp: time.Now(), Close: 20},
		{Timestamp: time.Now(), Close: 50},
	}

	tr := &ClipTransform{
		Field: "close",
		Min:   10,
		Max:   1000, // High max, won't clip upper
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// First value clipped to 10, others unchanged
	if result[0].Close != 10 {
		t.Errorf("candle[0].Close = %f, want 10 (clipped)", result[0].Close)
	}
	if result[1].Close != 20 {
		t.Errorf("candle[1].Close = %f, want 20 (unchanged)", result[1].Close)
	}
	if result[2].Close != 50 {
		t.Errorf("candle[2].Close = %f, want 50 (unchanged)", result[2].Close)
	}
}

func TestClipTransform_OnlyUpperBound(t *testing.T) {
	// Test clipping only at upper bound
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 50},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 200},
	}

	tr := &ClipTransform{
		Field: "close",
		Min:   0, // Low min, won't clip lower
		Max:   150,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// First two unchanged, last clipped to 150
	if result[0].Close != 50 {
		t.Errorf("candle[0].Close = %f, want 50 (unchanged)", result[0].Close)
	}
	if result[1].Close != 100 {
		t.Errorf("candle[1].Close = %f, want 100 (unchanged)", result[1].Close)
	}
	if result[2].Close != 150 {
		t.Errorf("candle[2].Close = %f, want 150 (clipped)", result[2].Close)
	}
}

func TestClipTransform_VolumeField(t *testing.T) {
	// Test clipping on volume field
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 50000, Volume: 50},   // Below min
		{Timestamp: time.Now(), Close: 50000, Volume: 500},  // In range
		{Timestamp: time.Now(), Close: 50000, Volume: 5000}, // Above max
	}

	tr := &ClipTransform{
		Field: "volume",
		Min:   100,
		Max:   1000,
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

	// Volume should be clipped
	expected := []float64{100, 500, 1000}
	for i, exp := range expected {
		if result[i].Volume != exp {
			t.Errorf("candle[%d].Volume = %f, want %f", i, result[i].Volume, exp)
		}
	}
}

func TestClipTransform_NegativeRange(t *testing.T) {
	// Test with negative range (e.g., for differenced data)
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: -50},
		{Timestamp: time.Now(), Close: 0},
		{Timestamp: time.Now(), Close: 50},
	}

	tr := &ClipTransform{
		Field: "close",
		Min:   -20,
		Max:   20,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	expected := []float64{-20, 0, 20}
	for i, exp := range expected {
		if result[i].Close != exp {
			t.Errorf("candle[%d].Close = %f, want %f", i, result[i].Close, exp)
		}
	}
}

func TestClipTransform_ExtremeOutliers(t *testing.T) {
	// Test with extreme outliers
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: -999999},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 999999},
	}

	tr := &ClipTransform{
		Field: "close",
		Min:   0,
		Max:   200,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Extreme values should be clipped
	if result[0].Close != 0 {
		t.Errorf("candle[0].Close = %f, want 0 (extreme low clipped)", result[0].Close)
	}
	if result[1].Close != 100 {
		t.Errorf("candle[1].Close = %f, want 100 (unchanged)", result[1].Close)
	}
	if result[2].Close != 200 {
		t.Errorf("candle[2].Close = %f, want 200 (extreme high clipped)", result[2].Close)
	}
}

func TestClipTransform_SingleCandle(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 150},
	}

	tr := &ClipTransform{
		Field: "close",
		Min:   0,
		Max:   100,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("result length = %d, want 1", len(result))
	}

	if result[0].Close != 100 {
		t.Errorf("candle[0].Close = %f, want 100 (clipped)", result[0].Close)
	}
}

func TestClipTransform_EmptyData(t *testing.T) {
	tr := &ClipTransform{
		Field: "close",
		Min:   0,
		Max:   100,
	}

	_, err := tr.Apply([]*market.Candle{})
	if err == nil {
		t.Error("Apply() with empty data should return error")
	}
}

func TestClipTransform_Type(t *testing.T) {
	tr := &ClipTransform{}
	if tr.Type() != "clip" {
		t.Errorf("Type() = %s, want clip", tr.Type())
	}
}

func TestClipTransform_Name(t *testing.T) {
	tr := &ClipTransform{
		Field: "close",
		Min:   10.5,
		Max:   99.9,
	}
	expected := "clip(field=close, min=10.50, max=99.90)"
	if tr.Name() != expected {
		t.Errorf("Name() = %s, want %s", tr.Name(), expected)
	}
}

func TestClipTransform_ZeroRange(t *testing.T) {
	// Test very narrow range
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 99},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 101},
	}

	tr := &ClipTransform{
		Field: "close",
		Min:   99.5,
		Max:   100.5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	expected := []float64{99.5, 100, 100.5}
	for i, exp := range expected {
		if result[i].Close != exp {
			t.Errorf("candle[%d].Close = %f, want %f", i, result[i].Close, exp)
		}
	}
}
