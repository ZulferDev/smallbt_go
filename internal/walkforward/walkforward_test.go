package walkforward

import (
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/analytics"
	"github.com/1jehuang/backtest/internal/backtest"
)

func TestWindowConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  WindowConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500},
			wantErr: false,
		},
		{
			name:    "zero train bars",
			config:  WindowConfig{TrainBars: 0, TestBars: 500, StepBars: 500},
			wantErr: true,
		},
		{
			name:    "negative train bars",
			config:  WindowConfig{TrainBars: -100, TestBars: 500, StepBars: 500},
			wantErr: true,
		},
		{
			name:    "zero test bars",
			config:  WindowConfig{TrainBars: 1000, TestBars: 0, StepBars: 500},
			wantErr: true,
		},
		{
			name:    "negative test bars",
			config:  WindowConfig{TrainBars: 1000, TestBars: -100, StepBars: 500},
			wantErr: true,
		},
		{
			name:    "zero step bars defaults to test bars",
			config:  WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If no error and StepBars was 0, verify it was set to TestBars
			if !tt.wantErr && tt.config.StepBars == 0 {
				if tt.config.StepBars != tt.config.TestBars {
					t.Errorf("StepBars not defaulted to TestBars")
				}
			}
		})
	}
}

func TestNewWalkForwardAnalysis(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, err := New(config)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if wfa == nil {
		t.Fatal("New() returned nil")
	}

	if len(wfa.Windows) != 0 {
		t.Errorf("Initial Windows should be empty, got %d", len(wfa.Windows))
	}

	if len(wfa.Results) != 0 {
		t.Errorf("Initial Results should be empty, got %d", len(wfa.Results))
	}
}

func TestGenerateWindows(t *testing.T) {
	tests := []struct {
		name         string
		config       WindowConfig
		totalBars    int
		wantWindows  int
		wantErr      bool
	}{
		{
			name:        "single window",
			config:      WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500},
			totalBars:   1500,
			wantWindows: 1,
			wantErr:     false,
		},
		{
			name:        "two non-overlapping windows",
			config:      WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500},
			totalBars:   2000,
			wantWindows: 2,
			wantErr:     false,
		},
		{
			name:        "overlapping windows",
			config:      WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 250},
			totalBars:   2000,
			wantWindows: 3,
			wantErr:     false,
		},
		{
			name:        "insufficient bars",
			config:      WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500},
			totalBars:   1000,
			wantWindows: 0,
			wantErr:     true,
		},
		{
			name:        "zero bars",
			config:      WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500},
			totalBars:   0,
			wantWindows: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wfa, _ := New(tt.config)
			err := wfa.GenerateWindows(tt.totalBars)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateWindows() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(wfa.Windows) != tt.wantWindows {
				t.Errorf("GenerateWindows() got %d windows, want %d", len(wfa.Windows), tt.wantWindows)
			}
		})
	}
}

func TestWindowBoundaries(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(2000)

	if len(wfa.Windows) != 2 {
		t.Fatalf("Expected 2 windows, got %d", len(wfa.Windows))
	}

	// Window 0
	w0 := wfa.Windows[0]
	if w0.TrainStart != 0 || w0.TrainEnd != 999 {
		t.Errorf("Window 0 train boundaries: got [%d, %d], want [0, 999]", w0.TrainStart, w0.TrainEnd)
	}
	if w0.TestStart != 1000 || w0.TestEnd != 1499 {
		t.Errorf("Window 0 test boundaries: got [%d, %d], want [1000, 1499]", w0.TestStart, w0.TestEnd)
	}

	// Window 1
	w1 := wfa.Windows[1]
	if w1.TrainStart != 500 || w1.TrainEnd != 1499 {
		t.Errorf("Window 1 train boundaries: got [%d, %d], want [500, 1499]", w1.TrainStart, w1.TrainEnd)
	}
	if w1.TestStart != 1500 || w1.TestEnd != 1999 {
		t.Errorf("Window 1 test boundaries: got [%d, %d], want [1500, 1999]", w1.TestStart, w1.TestEnd)
	}
}

func TestAddWindowResult(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(1500)

	// Create mock results
	trainResult := &backtest.BacktestResult{
		TotalTrades: 10,
		Metrics: &analytics.Metrics{
			TotalReturn: 25.0,
			SharpeRatio: 1.5,
		},
	}
	testResult := &backtest.BacktestResult{
		TotalTrades: 8,
		Metrics: &analytics.Metrics{
			TotalReturn: 20.0,
			SharpeRatio: 1.2,
		},
	}

	result := &WFWindowResult{
		TrainResult: trainResult,
		TestResult:  testResult,
	}

	err := wfa.AddWindowResult(0, result)
	if err != nil {
		t.Fatalf("AddWindowResult() error = %v", err)
	}

	if len(wfa.Results) != 1 {
		t.Errorf("Results length: got %d, want 1", len(wfa.Results))
	}

	storedResult := wfa.Results[0]
	if storedResult.WindowID != 0 {
		t.Errorf("Result WindowID: got %d, want 0", storedResult.WindowID)
	}
}

func TestAddWindowResultOutOfBounds(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(1500)

	result := &WFWindowResult{}

	err := wfa.AddWindowResult(5, result) // Out of bounds
	if err != ErrInvalidWindowID {
		t.Errorf("AddWindowResult() error = %v, want ErrInvalidWindowID", err)
	}
}

func TestGetWindow(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(1500)

	window := wfa.GetWindow(0)
	if window == nil {
		t.Fatal("GetWindow() returned nil")
	}

	if window.WindowID != 0 {
		t.Errorf("Window ID: got %d, want 0", window.WindowID)
	}

	nilWindow := wfa.GetWindow(5)
	if nilWindow != nil {
		t.Error("GetWindow() should return nil for out of bounds")
	}
}

func TestWindowCount(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)

	if wfa.WindowCount() != 0 {
		t.Errorf("WindowCount() before generate: got %d, want 0", wfa.WindowCount())
	}

	wfa.GenerateWindows(2000)
	if wfa.WindowCount() != 2 {
		t.Errorf("WindowCount() after generate: got %d, want 2", wfa.WindowCount())
	}
}

func TestCompleteWindows(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(1500)

	if wfa.CompleteWindows() != 0 {
		t.Errorf("CompleteWindows() before adding results: got %d, want 0", wfa.CompleteWindows())
	}

	result := &WFWindowResult{
		TrainResult: &backtest.BacktestResult{},
		TestResult:  &backtest.BacktestResult{},
	}
	wfa.AddWindowResult(0, result)

	if wfa.CompleteWindows() != 1 {
		t.Errorf("CompleteWindows() after adding result: got %d, want 1", wfa.CompleteWindows())
	}
}

func TestComputeAggregateNoResults(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(1500)

	_, err := wfa.ComputeAggregate()
	if err != ErrNoResults {
		t.Errorf("ComputeAggregate() error = %v, want ErrNoResults", err)
	}
}

func TestComputeAggregateIncomplete(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(2000) // 2 windows

	// Add result for only 1 window
	result := &WFWindowResult{
		TrainResult: &backtest.BacktestResult{},
		TestResult:  &backtest.BacktestResult{},
	}
	wfa.AddWindowResult(0, result)

	_, err := wfa.ComputeAggregate()
	if err != ErrIncompleteWindows {
		t.Errorf("ComputeAggregate() error = %v, want ErrIncompleteWindows", err)
	}
}

func TestComputeAggregateBasic(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(1500)

	// Create mock results with equity points
	equity1 := []backtest.EquityPoint{
		{Timestamp: time.Now(), Equity: 10000, Cash: 5000, Drawdown: 0, Exposure: 0.5},
		{Timestamp: time.Now().Add(time.Hour), Equity: 10100, Cash: 5050, Drawdown: 0, Exposure: 0.5},
	}

	metrics := &analytics.Metrics{
		TotalReturn: 3.0,
		SharpeRatio: 1.5,
	}

	trainResult := &backtest.BacktestResult{
		TotalTrades: 10,
		EquityCurve: equity1,
		Metrics:     metrics,
	}

	testResult := &backtest.BacktestResult{
		TotalTrades: 8,
		EquityCurve: equity1,
		Metrics:     metrics,
	}

	result := &WFWindowResult{
		WindowID:    0,
		TrainResult: trainResult,
		TestResult:  testResult,
	}

	wfa.AddWindowResult(0, result)

	agg, err := wfa.ComputeAggregate()
	if err != nil {
		t.Fatalf("ComputeAggregate() error = %v", err)
	}

	if agg == nil {
		t.Fatal("ComputeAggregate() returned nil")
	}

	if agg.WindowCount != 1 {
		t.Errorf("WindowCount: got %d, want 1", agg.WindowCount)
	}

	if agg.TotalTrades != 8 { // Out-of-sample trades
		t.Errorf("TotalTrades: got %d, want 8", agg.TotalTrades)
	}

	if agg.TotalReturn != 3.0 {
		t.Errorf("TotalReturn: got %f, want 3.0", agg.TotalReturn)
	}
}

func TestExportToJSONEmpty(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)

	_, err := wfa.ExportToJSON()
	if err != ErrNoResults {
		t.Errorf("ExportToJSON() error = %v, want ErrNoResults", err)
	}
}

func TestExportToCSVEmpty(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)

	_, err := wfa.ExportToCSV()
	if err != ErrNoResults {
		t.Errorf("ExportToCSV() error = %v, want ErrNoResults", err)
	}
}

func TestReportEmpty(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)

	report := wfa.Report()
	if report != "Walk Forward Analysis: No results available" {
		t.Errorf("Report() empty: got unexpected message")
	}
}

func TestMultipleWindows(t *testing.T) {
	config := WindowConfig{TrainBars: 1000, TestBars: 500, StepBars: 500}
	wfa, _ := New(config)
	wfa.GenerateWindows(3000) // 5 windows possible

	if wfa.WindowCount() < 2 {
		t.Fatalf("Expected at least 2 windows, got %d", wfa.WindowCount())
	}

	metrics := &analytics.Metrics{
		TotalReturn: 3.0,
		SharpeRatio: 1.5,
	}

	// Add results for all windows
	for i := 0; i < wfa.WindowCount(); i++ {
		trainResult := &backtest.BacktestResult{
			TotalTrades: 10,
			EquityCurve: []backtest.EquityPoint{
				{Timestamp: time.Now(), Equity: 10000, Cash: 5000, Drawdown: 0, Exposure: 0.5},
				{Timestamp: time.Now().Add(time.Hour), Equity: 10100, Cash: 5050, Drawdown: 0, Exposure: 0.5},
			},
			Metrics: metrics,
		}

		testResult := &backtest.BacktestResult{
			TotalTrades: 8,
			EquityCurve: []backtest.EquityPoint{
				{Timestamp: time.Now(), Equity: 10100, Cash: 5050, Drawdown: 0, Exposure: 0.5},
				{Timestamp: time.Now().Add(time.Hour), Equity: 10103, Cash: 5051, Drawdown: 0, Exposure: 0.5},
			},
			Metrics: metrics,
		}

		result := &WFWindowResult{
			TrainResult: trainResult,
			TestResult:  testResult,
		}

		wfa.AddWindowResult(i, result)
	}

	if wfa.CompleteWindows() != wfa.WindowCount() {
		t.Errorf("CompleteWindows: got %d, want %d", wfa.CompleteWindows(), wfa.WindowCount())
	}

	agg, err := wfa.ComputeAggregate()
	if err != nil {
		t.Fatalf("ComputeAggregate() error = %v", err)
	}

	if agg.WindowCount != wfa.WindowCount() {
		t.Errorf("Aggregate WindowCount: got %d, want %d", agg.WindowCount, wfa.WindowCount())
	}
}
