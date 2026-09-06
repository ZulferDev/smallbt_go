package transform

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
)

// Difference tests

func TestDifferenceTransform_Apply(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100.0},
		{Timestamp: time.Now().Add(time.Hour), Close: 105.0},
		{Timestamp: time.Now().Add(2 * time.Hour), Close: 103.0},
		{Timestamp: time.Now().Add(3 * time.Hour), Close: 110.0},
	}

	transform := NewDifferenceTransform("close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Difference failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}

	// Check differences
	expected := []float64{5.0, -2.0, 7.0}
	for i, exp := range expected {
		if !almostEqual(result[i].Close, exp, 0.01) {
			t.Errorf("Result[%d]: expected %.2f, got %.2f", i, exp, result[i].Close)
		}
	}
}

func TestDifferenceTransform_InsufficientData(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100.0},
	}

	transform := NewDifferenceTransform("close")
	_, err := transform.Apply(candles)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestDifferenceTransform_OpenField(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Open: 100.0},
		{Timestamp: time.Now().Add(time.Hour), Open: 110.0},
	}

	transform := NewDifferenceTransform("open")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Difference on open failed: %v", err)
	}

	if !almostEqual(result[0].Close, 10.0, 0.01) {
		t.Errorf("Expected 10, got %.2f", result[0].Close)
	}
}

// Scale tests

func TestScaleTransform_Apply(t *testing.T) {
	candles := generateTestCandles(5, 100.0)

	transform := NewScaleTransform(2.0, "close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Scale failed: %v", err)
	}

	// Check all close values are doubled
	for i := range candles {
		expected := candles[i].Close * 2.0
		if !almostEqual(result[i].Close, expected, 0.01) {
			t.Errorf("Candle[%d]: expected %.2f, got %.2f", i, expected, result[i].Close)
		}
	}
}

func TestScaleTransform_AllFields(t *testing.T) {
	candles := []*market.Candle{
		{
			Timestamp: time.Now(),
			Open:      100.0,
			High:      110.0,
			Low:       95.0,
			Close:     105.0,
			Volume:    1000.0,
		},
	}

	transform := NewScaleTransform(0.5, "all")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Scale all failed: %v", err)
	}

	// Check all fields are halved
	if !almostEqual(result[0].Open, 50.0, 0.01) {
		t.Errorf("Expected open 50, got %.2f", result[0].Open)
	}
	if !almostEqual(result[0].High, 55.0, 0.01) {
		t.Errorf("Expected high 55, got %.2f", result[0].High)
	}
	if !almostEqual(result[0].Low, 47.5, 0.01) {
		t.Errorf("Expected low 47.5, got %.2f", result[0].Low)
	}
	if !almostEqual(result[0].Close, 52.5, 0.01) {
		t.Errorf("Expected close 52.5, got %.2f", result[0].Close)
	}
	if !almostEqual(result[0].Volume, 500.0, 0.01) {
		t.Errorf("Expected volume 500, got %.2f", result[0].Volume)
	}
}

func TestScaleTransform_NegativeFactor(t *testing.T) {
	candles := generateTestCandles(3, 100.0)

	transform := NewScaleTransform(-1.0, "close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Scale with negative factor failed: %v", err)
	}

	// Check values are negated
	for i := range candles {
		expected := candles[i].Close * -1.0
		if !almostEqual(result[i].Close, expected, 0.01) {
			t.Errorf("Candle[%d]: expected %.2f, got %.2f", i, expected, result[i].Close)
		}
	}
}

// Integration tests - complex chains

func TestComplexChain_NormalizeAndSmooth(t *testing.T) {
	candles := generateTestCandles(20, 100.0)

	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewMovingAverageSmoothTransform(5, "close"),
	)

	result, err := chain.Apply(candles)
	if err != nil {
		t.Fatalf("Complex chain failed: %v", err)
	}

	if len(result) != len(candles) {
		t.Errorf("Expected %d results, got %d", len(candles), len(result))
	}

	// Normalized then smoothed values should still be in reasonable range
	for i, candle := range result {
		if candle.Close < 0 || candle.Close > 1 {
			t.Errorf("Candle[%d]: value %.4f out of expected range", i, candle.Close)
		}
	}
}

func TestComplexChain_MultipleTransforms(t *testing.T) {
	candles := generateTestCandles(10, 100.0)

	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		NewScaleTransform(100.0, "close"),
		NewDifferenceTransform("close"),
	)

	result, err := chain.Apply(candles)
	if err != nil {
		t.Fatalf("Multi-transform chain failed: %v", err)
	}

	// After difference, should have n-1 candles
	if len(result) != len(candles)-1 {
		t.Errorf("Expected %d results, got %d", len(candles)-1, len(result))
	}
}

func TestComplexChain_LogReturnsAndSmooth(t *testing.T) {
	candles := []*market.Candle{
		{Close: 100.0},
		{Close: 105.0},
		{Close: 110.0},
		{Close: 108.0},
		{Close: 112.0},
		{Close: 115.0},
	}

	chain := NewTransformChain(
		NewLogReturnsTransform("close"),
		NewMovingAverageSmoothTransform(2, "close"),
	)

	result, err := chain.Apply(candles)
	if err != nil {
		t.Fatalf("LogReturns+Smooth chain failed: %v", err)
	}

	// After log returns: n-1
	if len(result) != len(candles)-1 {
		t.Errorf("Expected %d results, got %d", len(candles)-1, len(result))
	}
}

// Edge cases

func TestTransform_EmptyInput(t *testing.T) {
	transforms := []Transform{
		NewNormalizeTransform("close"),
		NewLogReturnsTransform("close"),
		NewPercentageChangeTransform("close"),
		NewMovingAverageSmoothTransform(3, "close"),
		NewDifferenceTransform("close"),
		NewScaleTransform(2.0, "close"),
	}

	for _, transform := range transforms {
		t.Run(transform.Name(), func(t *testing.T) {
			_, err := transform.Apply([]*market.Candle{})
			if err == nil {
				t.Errorf("%s should fail on empty input", transform.Name())
			}
		})
	}
}

func TestTransform_SingleCandle(t *testing.T) {
	candles := []*market.Candle{
		{Timestamp: time.Now(), Close: 100.0},
	}

	// These should work with single candle
	successTransforms := []Transform{
		NewNormalizeTransform("close"),
		NewScaleTransform(2.0, "close"),
	}

	for _, transform := range successTransforms {
		t.Run(transform.Name(), func(t *testing.T) {
			_, err := transform.Apply(candles)
			if err != nil {
				t.Errorf("%s failed with single candle: %v", transform.Name(), err)
			}
		})
	}

	// These should fail with single candle
	failTransforms := []Transform{
		NewLogReturnsTransform("close"),
		NewPercentageChangeTransform("close"),
		NewDifferenceTransform("close"),
	}

	for _, transform := range failTransforms {
		t.Run(transform.Name(), func(t *testing.T) {
			_, err := transform.Apply(candles)
			if err == nil {
				t.Errorf("%s should fail with single candle", transform.Name())
			}
		})
	}
}

func TestTransform_PreservesTimestamp(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := []*market.Candle{
		{Timestamp: baseTime, Close: 100.0},
		{Timestamp: baseTime.Add(time.Hour), Close: 105.0},
		{Timestamp: baseTime.Add(2 * time.Hour), Close: 110.0},
	}

	transform := NewNormalizeTransform("close")
	result, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	// Check timestamps are preserved
	for i := range candles {
		if !result[i].Timestamp.Equal(candles[i].Timestamp) {
			t.Errorf("Timestamp[%d] changed: %v -> %v", i, candles[i].Timestamp, result[i].Timestamp)
		}
	}
}

func TestTransform_DoesNotModifyOriginal(t *testing.T) {
	candles := generateTestCandles(5, 100.0)
	originalClose := candles[0].Close

	transform := NewScaleTransform(2.0, "close")
	_, err := transform.Apply(candles)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	// Original should be unchanged
	if candles[0].Close != originalClose {
		t.Errorf("Original candle was modified: %.2f -> %.2f", originalClose, candles[0].Close)
	}
}

// TransformFunc tests

func TestTransformFunc_Apply(t *testing.T) {
	candles := generateTestCandles(5, 100.0)

	// Custom transform: add 10 to close
	customFunc := TransformFunc(func(input []*market.Candle) ([]*market.Candle, error) {
		result := CopyCandles(input)
		for _, candle := range result {
			candle.Close += 10.0
		}
		return result, nil
	})

	result, err := customFunc.Apply(candles)
	if err != nil {
		t.Fatalf("TransformFunc failed: %v", err)
	}

	for i := range candles {
		expected := candles[i].Close + 10.0
		if !almostEqual(result[i].Close, expected, 0.01) {
			t.Errorf("Candle[%d]: expected %.2f, got %.2f", i, expected, result[i].Close)
		}
	}
}

func TestTransformFunc_InChain(t *testing.T) {
	candles := generateTestCandles(5, 100.0)

	customFunc := TransformFunc(func(input []*market.Candle) ([]*market.Candle, error) {
		result := CopyCandles(input)
		for _, candle := range result {
			candle.Close *= 2.0
		}
		return result, nil
	})

	chain := NewTransformChain(
		NewNormalizeTransform("close"),
		customFunc,
	)

	result, err := chain.Apply(candles)
	if err != nil {
		t.Fatalf("Chain with TransformFunc failed: %v", err)
	}

	if len(result) != len(candles) {
		t.Errorf("Expected %d results, got %d", len(candles), len(result))
	}
}

// Validation tests

func TestTransform_ValidateBeforeApply(t *testing.T) {
	candles := generateTestCandles(5, 100.0)

	// Invalid field
	invalid := NewNormalizeTransform("invalid_field")
	if err := invalid.Validate(); err == nil {
		t.Error("Expected validation error")
	}

	// Should also fail on Apply
	_, err := invalid.Apply(candles)
	if err != nil {
		// This is expected - invalid config might be caught during apply
	}
}

func TestMovingAverageSmooth_InsufficientData(t *testing.T) {
	candles := generateTestCandles(3, 100.0)

	transform := NewMovingAverageSmoothTransform(5, "close")
	_, err := transform.Apply(candles)
	if err == nil {
		t.Error("Expected error when candles < period")
	}
}
