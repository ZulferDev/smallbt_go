package transform

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestDifferencingTransform_Validate(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		order   int
		wantErr bool
	}{
		{
			name:    "valid first order close",
			field:   "close",
			order:   1,
			wantErr: false,
		},
		{
			name:    "valid second order",
			field:   "close",
			order:   2,
			wantErr: false,
		},
		{
			name:    "valid third order",
			field:   "close",
			order:   3,
			wantErr: false,
		},
		{
			name:    "valid volume",
			field:   "volume",
			order:   1,
			wantErr: false,
		},
		{
			name:    "missing field",
			field:   "",
			order:   1,
			wantErr: true,
		},
		{
			name:    "invalid field",
			field:   "invalid",
			order:   1,
			wantErr: true,
		},
		{
			name:    "order zero",
			field:   "close",
			order:   0,
			wantErr: true,
		},
		{
			name:    "order too high",
			field:   "close",
			order:   4,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &DifferencingTransform{
				Field: tt.field,
				Order: tt.order,
			}
			err := tr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDifferencingTransform_FirstOrder(t *testing.T) {
	// Test first-order differencing: price changes
	// Prices: 100, 105, 103, 108, 110
	// Diffs:  0, 5, -2, 5, 2
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 105},
		{Timestamp: time.Now(), Close: 103},
		{Timestamp: time.Now(), Close: 108},
		{Timestamp: time.Now(), Close: 110},
	}

	tr := &DifferencingTransform{
		Field: "close",
		Order: 1,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	expected := []float64{0, 5, -2, 5, 2}
	for i, exp := range expected {
		if result[i].Close != exp {
			t.Errorf("candle[%d].Close = %f, want %f", i, result[i].Close, exp)
		}
	}
}

func TestDifferencingTransform_SecondOrder(t *testing.T) {
	// Test second-order differencing: acceleration
	// Prices:      100, 105, 103, 108, 110
	// 1st diff:    0,   5,   -2,  5,   2
	// 2nd diff:    0,   5,   -7,  7,   -3
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 105},
		{Timestamp: time.Now(), Close: 103},
		{Timestamp: time.Now(), Close: 108},
		{Timestamp: time.Now(), Close: 110},
	}

	tr := &DifferencingTransform{
		Field: "close",
		Order: 2,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	expected := []float64{0, 5, -7, 7, -3}
	for i, exp := range expected {
		if result[i].Close != exp {
			t.Errorf("candle[%d].Close = %f, want %f", i, result[i].Close, exp)
		}
	}
}

func TestDifferencingTransform_ThirdOrder(t *testing.T) {
	// Test third-order differencing
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 105},
		{Timestamp: time.Now(), Close: 103},
		{Timestamp: time.Now(), Close: 108},
		{Timestamp: time.Now(), Close: 110},
		{Timestamp: time.Now(), Close: 115},
	}

	tr := &DifferencingTransform{
		Field: "close",
		Order: 3,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Just verify it runs without error and produces correct length
	if len(result) != len(candles) {
		t.Errorf("result length = %d, want %d", len(result), len(candles))
	}

	// First value should be 0
	if result[0].Close != 0 {
		t.Errorf("candle[0].Close = %f, want 0", result[0].Close)
	}
}

func TestDifferencingTransform_VolumeField(t *testing.T) {
	// Test differencing on volume field
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 50000, Volume: 100},
		{Timestamp: time.Now(), Close: 50000, Volume: 150},
		{Timestamp: time.Now(), Close: 50000, Volume: 120},
		{Timestamp: time.Now(), Close: 50000, Volume: 180},
	}

	tr := &DifferencingTransform{
		Field: "volume",
		Order: 1,
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

	// Volume differences: 0, 50, -30, 60
	expectedVol := []float64{0, 50, -30, 60}
	for i, exp := range expectedVol {
		if result[i].Volume != exp {
			t.Errorf("candle[%d].Volume = %f, want %f", i, result[i].Volume, exp)
		}
	}
}

func TestDifferencingTransform_ConstantValues(t *testing.T) {
	// Test with constant values - all differences should be 0
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 100},
	}

	tr := &DifferencingTransform{
		Field: "close",
		Order: 1,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// All values should be 0
	for i, c := range result {
		if c.Close != 0 {
			t.Errorf("candle[%d].Close = %f, want 0", i, c.Close)
		}
	}
}

func TestDifferencingTransform_TrendRemoval(t *testing.T) {
	// Test that differencing removes linear trend
	// Linear trend: 100, 110, 120, 130, 140 (slope = 10)
	// After differencing: 0, 10, 10, 10, 10
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 110},
		{Timestamp: time.Now(), Close: 120},
		{Timestamp: time.Now(), Close: 130},
		{Timestamp: time.Now(), Close: 140},
	}

	tr := &DifferencingTransform{
		Field: "close",
		Order: 1,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// After differencing, linear trend becomes constant
	for i := 1; i < len(result); i++ {
		if result[i].Close != 10 {
			t.Errorf("candle[%d].Close = %f, want 10 (constant after trend removal)", i, result[i].Close)
		}
	}
}

func TestDifferencingTransform_InsufficientData(t *testing.T) {
	tests := []struct {
		name    string
		candles []*market.Candle
		order   int
		wantErr bool
	}{
		{
			name: "order 1 needs 2 candles",
			candles: []*market.Candle{
				{Timestamp: time.Now(), Close: 100},
			},
			order:   1,
			wantErr: true,
		},
		{
			name: "order 2 needs 3 candles",
			candles: []*market.Candle{
				{Timestamp: time.Now(), Close: 100},
				{Timestamp: time.Now(), Close: 105},
			},
			order:   2,
			wantErr: true,
		},
		{
			name: "order 3 needs 4 candles",
			candles: []*market.Candle{
				{Timestamp: time.Now(), Close: 100},
				{Timestamp: time.Now(), Close: 105},
				{Timestamp: time.Now(), Close: 110},
			},
			order:   3,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &DifferencingTransform{
				Field: "close",
				Order: tt.order,
			}

			_, err := tr.Apply(tt.candles)
			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDifferencingTransform_EmptyData(t *testing.T) {
	tr := &DifferencingTransform{
		Field: "close",
		Order: 1,
	}

	_, err := tr.Apply([]*market.Candle{})
	if err == nil {
		t.Error("Apply() with empty data should return error")
	}
}

func TestDifferencingTransform_Type(t *testing.T) {
	tr := &DifferencingTransform{}
	if tr.Type() != "difference" {
		t.Errorf("Type() = %s, want difference", tr.Type())
	}
}

func TestDifferencingTransform_Name(t *testing.T) {
	tr := &DifferencingTransform{
		Field: "close",
		Order: 2,
	}
	expected := "difference(field=close, order=2)"
	if tr.Name() != expected {
		t.Errorf("Name() = %s, want %s", tr.Name(), expected)
	}
}
