package analytics

import (
	"errors"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// mockAnalyzer is a test implementation of CustomAnalyzer.
type mockAnalyzer struct {
	name   string
	result interface{}
	err    error
}

func (m *mockAnalyzer) Name() string {
	return m.name
}

func (m *mockAnalyzer) Calculate(input AnalysisInput) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// TestNewRegistry tests registry creation.
func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.analyzers == nil {
		t.Fatal("Registry analyzers map is nil")
	}
	if len(r.analyzers) != 0 {
		t.Fatalf("Expected empty registry, got %d analyzers", len(r.analyzers))
	}
}

// TestRegister tests successful analyzer registration.
func TestRegister(t *testing.T) {
	r := NewRegistry()
	analyzer := &mockAnalyzer{name: "test_analyzer", result: 42.0}

	err := r.Register(analyzer)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !r.Has("test_analyzer") {
		t.Fatal("Analyzer not found after registration")
	}

	retrieved := r.Get("test_analyzer")
	if retrieved != analyzer {
		t.Fatal("Retrieved analyzer is not the same instance")
	}
}

// TestRegisterNil tests registering nil analyzer.
func TestRegisterNil(t *testing.T) {
	r := NewRegistry()
	err := r.Register(nil)
	if err == nil {
		t.Fatal("Expected error when registering nil analyzer")
	}
}

// TestRegisterEmptyName tests registering analyzer with empty name.
func TestRegisterEmptyName(t *testing.T) {
	r := NewRegistry()
	analyzer := &mockAnalyzer{name: "", result: 1.0}

	err := r.Register(analyzer)
	if err == nil {
		t.Fatal("Expected error when registering analyzer with empty name")
	}
}

// TestRegisterDuplicate tests registering duplicate analyzer name.
func TestRegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	analyzer1 := &mockAnalyzer{name: "duplicate", result: 1.0}
	analyzer2 := &mockAnalyzer{name: "duplicate", result: 2.0}

	err := r.Register(analyzer1)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	err = r.Register(analyzer2)
	if err == nil {
		t.Fatal("Expected error when registering duplicate analyzer")
	}
}

// TestUnregister tests analyzer unregistration.
func TestUnregister(t *testing.T) {
	r := NewRegistry()
	analyzer := &mockAnalyzer{name: "test", result: 1.0}

	r.Register(analyzer)
	if !r.Has("test") {
		t.Fatal("Analyzer not registered")
	}

	r.Unregister("test")
	if r.Has("test") {
		t.Fatal("Analyzer still exists after unregister")
	}
}

// TestUnregisterNonExistent tests unregistering non-existent analyzer.
func TestUnregisterNonExistent(t *testing.T) {
	r := NewRegistry()
	// Should not panic
	r.Unregister("nonexistent")
}

// TestGet tests retrieving analyzer.
func TestGet(t *testing.T) {
	r := NewRegistry()
	analyzer := &mockAnalyzer{name: "test", result: 99.0}
	r.Register(analyzer)

	retrieved := r.Get("test")
	if retrieved == nil {
		t.Fatal("Get returned nil for registered analyzer")
	}
	if retrieved.Name() != "test" {
		t.Fatalf("Expected name 'test', got %q", retrieved.Name())
	}
}

// TestGetNonExistent tests retrieving non-existent analyzer.
func TestGetNonExistent(t *testing.T) {
	r := NewRegistry()
	retrieved := r.Get("nonexistent")
	if retrieved != nil {
		t.Fatal("Expected nil for non-existent analyzer")
	}
}

// TestHas tests checking analyzer existence.
func TestHas(t *testing.T) {
	r := NewRegistry()
	analyzer := &mockAnalyzer{name: "exists", result: 1.0}
	r.Register(analyzer)

	if !r.Has("exists") {
		t.Fatal("Has returned false for existing analyzer")
	}
	if r.Has("does_not_exist") {
		t.Fatal("Has returned true for non-existent analyzer")
	}
}

// TestList tests listing all analyzer names.
func TestList(t *testing.T) {
	r := NewRegistry()

	// Empty registry
	names := r.List()
	if len(names) != 0 {
		t.Fatalf("Expected empty list, got %d names", len(names))
	}

	// Register multiple analyzers
	r.Register(&mockAnalyzer{name: "analyzer1", result: 1.0})
	r.Register(&mockAnalyzer{name: "analyzer2", result: 2.0})
	r.Register(&mockAnalyzer{name: "analyzer3", result: 3.0})

	names = r.List()
	if len(names) != 3 {
		t.Fatalf("Expected 3 names, got %d", len(names))
	}

	// Check all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}
	if !nameMap["analyzer1"] || !nameMap["analyzer2"] || !nameMap["analyzer3"] {
		t.Fatal("Not all analyzer names are present in list")
	}
}

// TestClear tests clearing all analyzers.
func TestClear(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAnalyzer{name: "analyzer1", result: 1.0})
	r.Register(&mockAnalyzer{name: "analyzer2", result: 2.0})

	if len(r.List()) != 2 {
		t.Fatal("Expected 2 analyzers before clear")
	}

	r.Clear()

	if len(r.List()) != 0 {
		t.Fatalf("Expected 0 analyzers after clear, got %d", len(r.List()))
	}
}

// TestCalculate tests calculating a specific analyzer.
func TestCalculate(t *testing.T) {
	r := NewRegistry()
	analyzer := &mockAnalyzer{name: "test", result: 123.45}
	r.Register(analyzer)

	input := AnalysisInput{
		InitialCash: 10000,
		FinalEquity: 12000,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(24 * time.Hour),
	}

	result, err := r.Calculate("test", input)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if result != 123.45 {
		t.Fatalf("Expected result 123.45, got %v", result)
	}
}

// TestCalculateNonExistent tests calculating non-existent analyzer.
func TestCalculateNonExistent(t *testing.T) {
	r := NewRegistry()
	input := AnalysisInput{}

	_, err := r.Calculate("nonexistent", input)
	if err == nil {
		t.Fatal("Expected error when calculating non-existent analyzer")
	}
}

// TestCalculateWithError tests analyzer returning error.
func TestCalculateWithError(t *testing.T) {
	r := NewRegistry()
	expectedErr := errors.New("calculation failed")
	analyzer := &mockAnalyzer{name: "failing", err: expectedErr}
	r.Register(analyzer)

	input := AnalysisInput{}
	_, err := r.Calculate("failing", input)
	if err == nil {
		t.Fatal("Expected error from failing analyzer")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestCalculateAll tests calculating all analyzers.
func TestCalculateAll(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAnalyzer{name: "analyzer1", result: 100.0})
	r.Register(&mockAnalyzer{name: "analyzer2", result: 200.0})
	r.Register(&mockAnalyzer{name: "analyzer3", result: "custom"})

	input := AnalysisInput{
		InitialCash: 10000,
		FinalEquity: 12000,
	}

	results := r.CalculateAll(input)
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	if results["analyzer1"] != 100.0 {
		t.Fatalf("Expected analyzer1 result 100.0, got %v", results["analyzer1"])
	}
	if results["analyzer2"] != 200.0 {
		t.Fatalf("Expected analyzer2 result 200.0, got %v", results["analyzer2"])
	}
	if results["analyzer3"] != "custom" {
		t.Fatalf("Expected analyzer3 result 'custom', got %v", results["analyzer3"])
	}
}

// TestCalculateAllWithErrors tests CalculateAll with failing analyzers.
func TestCalculateAllWithErrors(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockAnalyzer{name: "success", result: 42.0})
	r.Register(&mockAnalyzer{name: "failure", err: errors.New("fail")})

	input := AnalysisInput{}
	results := r.CalculateAll(input)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Success analyzer should have result
	if results["success"] != 42.0 {
		t.Fatalf("Expected success result 42.0, got %v", results["success"])
	}

	// Failure analyzer should have error
	if _, ok := results["failure"].(error); !ok {
		t.Fatalf("Expected error result for failure analyzer, got %v", results["failure"])
	}
}

// TestGlobalRegistry tests package-level functions.
func TestGlobalRegistry(t *testing.T) {
	// Clear global registry before test
	globalRegistry.Clear()

	analyzer := &mockAnalyzer{name: "global_test", result: 99.0}

	// Test Register
	err := Register(analyzer)
	if err != nil {
		t.Fatalf("Global Register failed: %v", err)
	}

	// Test Has
	if !Has("global_test") {
		t.Fatal("Global Has returned false")
	}

	// Test Get
	retrieved := Get("global_test")
	if retrieved == nil {
		t.Fatal("Global Get returned nil")
	}

	// Test List
	names := List()
	found := false
	for _, name := range names {
		if name == "global_test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Analyzer not found in global List")
	}

	// Test Calculate
	input := AnalysisInput{}
	result, err := Calculate("global_test", input)
	if err != nil {
		t.Fatalf("Global Calculate failed: %v", err)
	}
	if result != 99.0 {
		t.Fatalf("Expected result 99.0, got %v", result)
	}

	// Test CalculateAll
	results := CalculateAll(input)
	if len(results) == 0 {
		t.Fatal("Global CalculateAll returned empty results")
	}

	// Test Unregister
	Unregister("global_test")
	if Has("global_test") {
		t.Fatal("Analyzer still exists after global Unregister")
	}

	// Clean up
	globalRegistry.Clear()
}

// TestGlobalRegistryInstance tests GlobalRegistry() function.
func TestGlobalRegistryInstance(t *testing.T) {
	registry := GlobalRegistry()
	if registry == nil {
		t.Fatal("GlobalRegistry() returned nil")
	}
	if registry != globalRegistry {
		t.Fatal("GlobalRegistry() returned different instance")
	}
}

// Example custom analyzer for documentation.
type WinStreakAnalyzer struct{}

func (w *WinStreakAnalyzer) Name() string {
	return "win_streak"
}

func (w *WinStreakAnalyzer) Calculate(input AnalysisInput) (interface{}, error) {
	if len(input.TradeHistory) == 0 {
		return 0, nil
	}

	maxStreak := 0
	currentStreak := 0

	for _, trade := range input.TradeHistory {
		if trade.NetPnL > 0 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
	}

	return maxStreak, nil
}

// TestRealWorldAnalyzer tests a realistic custom analyzer.
func TestRealWorldAnalyzer(t *testing.T) {
	r := NewRegistry()
	r.Register(&WinStreakAnalyzer{})

	input := AnalysisInput{
		TradeHistory: []portfolio.Trade{
			{NetPnL: 100}, // Win
			{NetPnL: 50},  // Win
			{NetPnL: -20}, // Loss
			{NetPnL: 30},  // Win
			{NetPnL: 40},  // Win
			{NetPnL: 25},  // Win (max streak = 3)
			{NetPnL: -10}, // Loss
		},
	}

	result, err := r.Calculate("win_streak", input)
	if err != nil {
		t.Fatalf("WinStreakAnalyzer failed: %v", err)
	}

	maxStreak, ok := result.(int)
	if !ok {
		t.Fatalf("Expected int result, got %T", result)
	}

	if maxStreak != 3 {
		t.Fatalf("Expected max streak 3, got %d", maxStreak)
	}
}
