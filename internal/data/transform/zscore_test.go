package transform

import (
	"math"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

func TestZScoreTransform_Validate(t *testing.T) {
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
			name:    "window zero",
			field:   "close",
			window:  0,
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
			tr := &ZScoreTransform{
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

func TestZScoreTransform_Apply(t *testing.T) {
	// Create simple test data with known statistics
	// Values: 100, 105, 110, 115, 120
	// Mean of last 5 = 110
	// Stddev = sqrt(((−10)² + (−5)² + 0² + 5² + 10²) / 5) = sqrt(250/5) = sqrt(50) ≈ 7.071
	// Z-score of 120 = (120 - 110) / 7.071 ≈ 1.414
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 105},
		{Timestamp: time.Now(), Close: 110},
		{Timestamp: time.Now(), Close: 115},
		{Timestamp: time.Now(), Close: 120},
	}

	tr := &ZScoreTransform{
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

	// First 4 candles should be 0 (not enough data)
	for i := 0; i < 4; i++ {
		if result[i].Close != 0 {
			t.Errorf("candle[%d].Close = %f, want 0", i, result[i].Close)
		}
	}

	// Last candle should have z-score ≈ 1.414
	expectedZ := 1.414
	tolerance := 0.01
	actualZ := result[4].Close
	if math.Abs(actualZ-expectedZ) > tolerance {
		t.Errorf("candle[4].Close = %f, want ≈ %f", actualZ, expectedZ)
	}
}

func TestZScoreTransform_RollingWindow(t *testing.T) {
	// Test that window rolls correctly
	// Create data where z-score should change as window moves
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100}, // mean=100, z=0
		{Timestamp: time.Now(), Close: 100}, // mean=100, z=0
		{Timestamp: time.Now(), Close: 100}, // mean=100, z=0
		{Timestamp: time.Now(), Close: 200}, // mean=125, stddev≈43.3, z≈1.73
		{Timestamp: time.Now(), Close: 200}, // mean=140, stddev≈44.7, z≈1.34
	}

	tr := &ZScoreTransform{
		Field:  "close",
		Window: 3,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// First 2 should be 0
	for i := 0; i < 2; i++ {
		if result[i].Close != 0 {
			t.Errorf("candle[%d].Close = %f, want 0", i, result[i].Close)
		}
	}

	// candle[2]: all values are 100, stddev=0, z-score=0
	if result[2].Close != 0 {
		t.Errorf("candle[2].Close = %f, want 0 (constant values)", result[2].Close)
	}

	// candle[3]: values are [100,100,200], should have positive z-score
	if result[3].Close <= 0 {
		t.Errorf("candle[3].Close = %f, want positive z-score", result[3].Close)
	}

	// candle[4]: values are [100,200,200], z-score should be positive but different
	if result[4].Close <= 0 {
		t.Errorf("candle[4].Close = %f, want positive z-score", result[4].Close)
	}

	// candle[4] should have different z-score than candle[3]
	if result[3].Close == result[4].Close {
		t.Errorf("candle[3] and candle[4] should have different z-scores")
	}
}

func TestZScoreTransform_VolumeField(t *testing.T) {
	// Test z-score on volume field
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 50000, Volume: 100},
		{Timestamp: time.Now(), Close: 50000, Volume: 200},
		{Timestamp: time.Now(), Close: 50000, Volume: 300},
		{Timestamp: time.Now(), Close: 50000, Volume: 400},
		{Timestamp: time.Now(), Close: 50000, Volume: 500}, // mean=300, stddev≈141.4, z≈1.414
	}

	tr := &ZScoreTransform{
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

	// Volume should be z-scored
	expectedZ := 1.414
	tolerance := 0.01
	actualZ := result[4].Volume
	if math.Abs(actualZ-expectedZ) > tolerance {
		t.Errorf("candle[4].Volume = %f, want ≈ %f", actualZ, expectedZ)
	}
}

func TestZScoreTransform_InsufficientData(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 105},
	}

	tr := &ZScoreTransform{
		Field:  "close",
		Window: 5,
	}

	_, err := tr.Apply(candles)
	if err == nil {
		t.Error("Apply() with insufficient data should return error")
	}
}

func TestZScoreTransform_EmptyData(t *testing.T) {
	tr := &ZScoreTransform{
		Field:  "close",
		Window: 5,
	}

	_, err := tr.Apply([]*market.Candle{})
	if err == nil {
		t.Error("Apply() with empty data should return error")
	}
}

func TestZScoreTransform_OutlierDetection(t *testing.T) {
	// Test that outliers produce high z-scores
	// Normal values around 100, then a spike to 300
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100},
		{Timestamp: time.Now(), Close: 105},
		{Timestamp: time.Now(), Close: 102},
		{Timestamp: time.Now(), Close: 98},
		{Timestamp: time.Now(), Close: 300}, // Outlier
	}

	tr := &ZScoreTransform{
		Field:  "close",
		Window: 5,
	}

	result, err := tr.Apply(candles)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Last value should have high z-score (typically > 2 or 3 for outliers)
	zScore := result[4].Close
	if zScore < 1.9 {
		t.Errorf("outlier z-score = %f, want > 1.9", zScore)
	}
}

func TestZScoreTransform_Type(t *testing.T) {
	tr := &ZScoreTransform{}
	if tr.Type() != "zscore" {
		t.Errorf("Type() = %s, want zscore", tr.Type())
	}
}

func TestCalculateMean(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{
			name:   "simple average",
			values: []float64{1, 2, 3, 4, 5},
			want:   3.0,
		},
		{
			name:   "all same",
			values: []float64{10, 10, 10, 10},
			want:   10.0,
		},
		{
			name:   "negative values",
			values: []float64{-5, 0, 5},
			want:   0.0,
		},
		{
			name:   "empty",
			values: []float64{},
			want:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMean(tt.values)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("calculateMean() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestCalculateStdDev(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		mean   float64
		want   float64
	}{
		{
			name:   "simple case",
			values: []float64{2, 4, 4, 4, 5, 5, 7, 9},
			mean:   5.0,
			want:   2.0,
		},
		{
			name:   "all same (zero stddev)",
			values: []float64{5, 5, 5, 5},
			mean:   5.0,
			want:   0.0,
		},
		{
			name:   "empty",
			values: []float64{},
			mean:   0.0,
			want:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateStdDev(tt.values, tt.mean)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("calculateStdDev() = %f, want %f", got, tt.want)
			}
		})
	}
}
