package transform

import (
	"math"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Test helpers

func generateTestCandles(count int, basePrice float64) []*market.Candle {
	candles := make([]*market.Candle, count)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		price := basePrice + float64(i)
		candles[i] = &market.Candle{
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour),
			Open:      price,
			High:      price + 2,
			Low:       price - 1,
			Close:     price + 1,
			Volume:    1000.0 + float64(i*10),
		}
	}

	return candles
}

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// TransformChain tests

func TestNewTransformChain(t *testing.T) {
	t1 := NewNormalizeTransform("close")
	t2 := NewScaleTransform(2.0, "close")

	chain := NewTransformChain(t1, t2)

	if chain.Len() != 2 {
		t.Errorf("Expected chain length 2, got %d", chain.Len())
	}
}

func TestTransformChain_Apply(t *testing.T) {
	candles := generateTestCandles(10, 100.0)

	// Chain: normalize then scale by 10
	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewScaleTransform(10.0, "close"),
	)

	result, err := chain.Apply(candles)
	if err != nil {
		t.Fatalf("Chain.Apply failed: %v", err)
	}

	if len(result) != len(candles) {
		t.Errorf("Expected %d candles, got %d", len(candles), len(result))
	}

	// First candle should be normalized to 0, then scaled to 0
	if !almostEqual(result[0].Close, 0.0, 0.01) {
		t.Errorf("Expected first close ~0, got %.4f", result[0].Close)
	}

	// Last candle should be normalized to 1, then scaled to 10
	if !almostEqual(result[len(result)-1].Close, 10.0, 0.01) {
		t.Errorf("Expected last close ~10, got %.4f", result[len(result)-1].Close)
	}
}

func TestTransformChain_EmptyChain(t *testing.T) {
	candles := generateTestCandles(5, 100.0)
	chain := NewTransformChain()

	result, err := chain.Apply(candles)
	if err != nil {
		t.Fatalf("Empty chain failed: %v", err)
	}

	if len(result) != len(candles) {
		t.Errorf("Expected %d candles, got %d", len(candles), len(result))
	}
}

func TestTransformChain_Add(t *testing.T) {
	chain := NewTransformChain()
	if chain.Len() != 0 {
		t.Errorf("Expected length 0, got %d", chain.Len())
	}

	chain.Add(NewNormalizeTransform("close"))
	if chain.Len() != 1 {
		t.Errorf("Expected length 1, got %d", chain.Len())
	}

	chain.Add(NewScaleTransform(2.0, "close"))
	if chain.Len() != 2 {
		t.Errorf("Expected length 2, got %d", chain.Len())
	}
}

func TestTransformChain_Get(t *testing.T) {
	t1 := NewNormalizeTransform("close")
	t2 := NewScaleTransform(2.0, "close")
	chain := NewTransformChain(t1, t2)

	if chain.Get(0) != t1 {
		t.Error("Get(0) did not return first transform")
	}

	if chain.Get(1) != t2 {
		t.Error("Get(1) did not return second transform")
	}

	if chain.Get(2) != nil {
		t.Error("Get(2) should return nil for out of bounds")
	}

	if chain.Get(-1) != nil {
		t.Error("Get(-1) should return nil for negative index")
	}
}

func TestTransformChain_Clear(t *testing.T) {
	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewScaleTransform(2.0, "close"),
	)

	if chain.Len() != 2 {
		t.Errorf("Expected length 2, got %d", chain.Len())
	}

	chain.Clear()

	if chain.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", chain.Len())
	}
}

func TestTransformChain_Clone(t *testing.T) {
	original := NewTransformChain(
		NewNormalizeTransform("close"),
		NewScaleTransform(2.0, "close"),
	)

	clone := original.Clone()

	if clone.Len() != original.Len() {
		t.Errorf("Clone length %d != original length %d", clone.Len(), original.Len())
	}

	// Modify clone
	clone.Add(NewDifferenceTransform("close"))

	if clone.Len() == original.Len() {
		t.Error("Modifying clone affected original")
	}
}

func TestTransformChain_Validate(t *testing.T) {
	// Valid chain
	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewScaleTransform(2.0, "close"),
	)

	if err := chain.Validate(); err != nil {
		t.Errorf("Valid chain validation failed: %v", err)
	}

	// Invalid chain
	invalidChain := NewTransformChain(
		NewMovingAverageSmoothTransform(0, "close"), // Invalid period
	)

	if err := invalidChain.Validate(); err == nil {
		t.Error("Expected validation error for invalid transform")
	}
}

// Normalize tests

func TestNormalizeTransform_Apply(t *testing.T) {
	candles := generateTestCandles(10, 100.0)

	transform := NewNormalizeTransform("close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	// Check normalization
	if !almostEqual(result[0].Close, 0.0, 0.01) {
		t.Errorf("Expected first close ~0, got %.4f", result[0].Close)
	}

	if !almostEqual(result[len(result)-1].Close, 1.0, 0.01) {
		t.Errorf("Expected last close ~1, got %.4f", result[len(result)-1].Close)
	}

	// Check all values are in [0, 1]
	for i, candle := range result {
		if candle.Close < 0 || candle.Close > 1 {
			t.Errorf("Candle %d close %.4f out of range [0, 1]", i, candle.Close)
		}
	}
}

func TestNormalizeTransform_AllFields(t *testing.T) {
	candles := generateTestCandles(10, 100.0)

	transform := NewNormalizeTransform("all")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Normalize all failed: %v", err)
	}

	// Check all OHLC fields are normalized
	for i, candle := range result {
		for _, value := range []float64{candle.Open, candle.High, candle.Low, candle.Close} {
			if value < 0 || value > 1 {
				t.Errorf("Candle %d has value %.4f out of range [0, 1]", i, value)
			}
		}
	}
}

func TestNormalizeTransform_ConstantValues(t *testing.T) {
	// All same values
	candles := make([]*market.Candle, 5)
	for i := range candles {
		candles[i] = &market.Candle{
			Timestamp: time.Now(),
			Open:      100.0,
			High:      100.0,
			Low:       100.0,
			Close:     100.0,
			Volume:    1000.0,
		}
	}

	transform := NewNormalizeTransform("close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Normalize constant failed: %v", err)
	}

	// Should normalize to 0.5
	for _, candle := range result {
		if !almostEqual(candle.Close, 0.5, 0.01) {
			t.Errorf("Expected close 0.5, got %.4f", candle.Close)
		}
	}
}

func TestNormalizeTransform_Validate(t *testing.T) {
	tests := []struct {
		field   string
		wantErr bool
	}{
		{"close", false},
		{"open", false},
		{"high", false},
		{"low", false},
		{"volume", false},
		{"all", false},
		{"", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			transform := NewNormalizeTransform(tt.field)
			err := transform.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// LogReturns tests

func TestLogReturnsTransform_Apply(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100.0},
		{Timestamp: time.Now().Add(time.Hour), Close: 110.0},
		{Timestamp: time.Now().Add(2 * time.Hour), Close: 121.0},
	}

	transform := NewLogReturnsTransform("close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("LogReturns failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}

	// ln(110/100) = ln(1.1) ≈ 0.0953
	expected := math.Log(1.1)
	if !almostEqual(result[0].Close, expected, 0.001) {
		t.Errorf("Expected %.4f, got %.4f", expected, result[0].Close)
	}
}

func TestLogReturnsTransform_InsufficientData(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100.0},
	}

	transform := NewLogReturnsTransform("close")
	_, err := transform.Apply(candles)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestLogReturnsTransform_NegativePrice(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100.0},
		{Timestamp: time.Now().Add(time.Hour), Close: -10.0},
	}

	transform := NewLogReturnsTransform("close")
	_, err := transform.Apply(candles)
	if err == nil {
		t.Error("Expected error for negative price")
	}
}

// PercentageChange tests

func TestPercentageChangeTransform_Apply(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100.0},
		{Timestamp: time.Now().Add(time.Hour), Close: 110.0},
		{Timestamp: time.Now().Add(2 * time.Hour), Close: 99.0},
	}

	transform := NewPercentageChangeTransform("close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("PercentageChange failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}

	// (110-100)/100 * 100 = 10%
	if !almostEqual(result[0].Close, 10.0, 0.01) {
		t.Errorf("Expected 10%%, got %.2f%%", result[0].Close)
	}

	// (99-110)/110 * 100 = -10%
	expected := ((99.0 - 110.0) / 110.0) * 100
	if !almostEqual(result[1].Close, expected, 0.01) {
		t.Errorf("Expected %.2f%%, got %.2f%%", expected, result[1].Close)
	}
}

func TestPercentageChangeTransform_ZeroPrice(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 0.0},
		{Timestamp: time.Now().Add(time.Hour), Close: 100.0},
	}

	transform := NewPercentageChangeTransform("close")
	_, err := transform.Apply(candles)
	if err == nil {
		t.Error("Expected error for zero price")
	}
}

// MovingAverageSmooth tests

func TestMovingAverageSmoothTransform_Apply(t *testing.T) {
	candles := []*market.Candle{
		{Close: 10.0},
		{Close: 20.0},
		{Close: 30.0},
		{Close: 40.0},
		{Close: 50.0},
	}

	transform := NewMovingAverageSmoothTransform(3, "close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("MASmooth failed: %v", err)
	}

	if len(result) != len(candles) {
		t.Errorf("Expected %d results, got %d", len(candles), len(result))
	}

	// MA(3) at index 2: (10+20+30)/3 = 20
	if !almostEqual(result[2].Close, 20.0, 0.01) {
		t.Errorf("Expected 20, got %.2f", result[2].Close)
	}

	// MA(3) at index 4: (30+40+50)/3 = 40
	if !almostEqual(result[4].Close, 40.0, 0.01) {
		t.Errorf("Expected 40, got %.2f", result[4].Close)
	}
}

func TestMovingAverageSmoothTransform_InvalidPeriod(t *testing.T) {
	transform := NewMovingAverageSmoothTransform(1, "close")
	if err := transform.Validate(); err == nil {
		t.Error("Expected validation error for period < 2")
	}
}

// Helper function tests

func TestValidateCandles(t *testing.T) {
	tests := []struct {
		name    string
		candles []*market.Candle
		wantErr bool
	}{
		{
			name:    "valid",
			candles: generateTestCandles(5, 100.0),
			wantErr: false,
		},
		{
			name:    "empty",
			candles: []*market.Candle{},
			wantErr: true,
		},
		{
			name:    "nil candle",
			candles: []*market.Candle{nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCandles(tt.candles)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCandles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCopyCandles(t *testing.T) {
	original := generateTestCandles(5, 100.0)
	copied := CopyCandles(original)

	if len(copied) != len(original) {
		t.Errorf("Copy length %d != original length %d", len(copied), len(original))
	}

	// Modify copy
	copied[0].Close = 999.0

	// Original should be unchanged
	if original[0].Close == 999.0 {
		t.Error("Modifying copy affected original")
	}
}
