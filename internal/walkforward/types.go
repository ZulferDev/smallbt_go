package walkforward

import (
	"time"

	"github.com/1jehuang/backtest/internal/backtest"
)

// WindowConfig defines training and testing window sizes.
type WindowConfig struct {
	// TrainBars is the number of bars for training period.
	TrainBars int

	// TestBars is the number of bars for testing (out-of-sample) period.
	TestBars int

	// StepBars is the number of bars to step forward for the next window.
	// If 0, defaults to TestBars (no overlap).
	StepBars int
}

// Validate checks window configuration is valid.
func (wc *WindowConfig) Validate() error {
	if wc.TrainBars <= 0 {
		return ErrInvalidTrainBars
	}
	if wc.TestBars <= 0 {
		return ErrInvalidTestBars
	}
	if wc.StepBars == 0 {
		wc.StepBars = wc.TestBars
	}
	return nil
}

// Window represents a single train/test period.
type Window struct {
	// WindowID is the unique identifier for this window (0-indexed).
	WindowID int

	// TrainStart is the index of the first training bar.
	TrainStart int

	// TrainEnd is the index of the last training bar (inclusive).
	TrainEnd int

	// TestStart is the index of the first test bar.
	TestStart int

	// TestEnd is the index of the last test bar (inclusive).
	TestEnd int
}

// WalkForwardAnalysis performs rolling window backtesting.
type WalkForwardAnalysis struct {
	Config  WindowConfig
	Windows []Window

	// Results maps window ID to backtest results.
	Results map[int]*WFWindowResult

	// AggregateResult contains out-of-sample aggregate metrics.
	AggregateResult *WFAggregateResult
}

// WFWindowResult holds backtest results for a single window.
type WFWindowResult struct {
	WindowID int

	// Training period results
	TrainResult *backtest.BacktestResult
	TrainStart  time.Time
	TrainEnd    time.Time

	// Testing (out-of-sample) period results
	TestResult *backtest.BacktestResult
	TestStart  time.Time
	TestEnd    time.Time

	// Parameter values used for this window
	Parameters map[string]float64
}

// WFAggregateResult aggregates out-of-sample performance across all windows.
type WFAggregateResult struct {
	// Total out-of-sample results
	TotalTrades        int
	TotalReturn        float64
	CAGR               float64
	SharpeRatio        float64
	SortinoRatio       float64
	MaxDrawdown        float64
	CalmarRatio        float64
	WinRate            float64
	ProfitFactor       float64
	Expectancy         float64
	AverageWin         float64
	AverageLoss        float64
	AverageTradeReturn float64

	// In-sample vs out-of-sample comparison
	InSampleAvgSharpe     float64
	OutOfSampleAvgSharpe  float64
	SharpeRatioDegradation float64 // (in-sample - out-of-sample) / in-sample

	// Window-level metrics
	WindowCount         int
	WindowResults       []WFWindowResult
	OutOfSampleEquity   []backtest.EquityPoint
}

// New creates a new WalkForwardAnalysis instance.
func New(config WindowConfig) (*WalkForwardAnalysis, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &WalkForwardAnalysis{
		Config:  config,
		Windows: []Window{},
		Results: make(map[int]*WFWindowResult),
	}, nil
}

// GenerateWindows creates rolling windows from total bar count.
func (wfa *WalkForwardAnalysis) GenerateWindows(totalBars int) error {
	if totalBars < wfa.Config.TrainBars+wfa.Config.TestBars {
		return ErrInsufficientBars
	}

	wfa.Windows = []Window{}
	windowID := 0

	for trainStart := 0; trainStart+wfa.Config.TrainBars+wfa.Config.TestBars <= totalBars; trainStart += wfa.Config.StepBars {
		window := Window{
			WindowID:  windowID,
			TrainStart: trainStart,
			TrainEnd:   trainStart + wfa.Config.TrainBars - 1,
			TestStart:  trainStart + wfa.Config.TrainBars,
			TestEnd:    trainStart + wfa.Config.TrainBars + wfa.Config.TestBars - 1,
		}
		wfa.Windows = append(wfa.Windows, window)
		windowID++
	}

	return nil
}

// AddWindowResult stores results for a window.
func (wfa *WalkForwardAnalysis) AddWindowResult(windowID int, result *WFWindowResult) error {
	if windowID < 0 || windowID >= len(wfa.Windows) {
		return ErrInvalidWindowID
	}

	result.WindowID = windowID
	wfa.Results[windowID] = result
	return nil
}

// GetWindow returns a window by ID.
func (wfa *WalkForwardAnalysis) GetWindow(windowID int) *Window {
	if windowID < 0 || windowID >= len(wfa.Windows) {
		return nil
	}
	return &wfa.Windows[windowID]
}

// WindowCount returns the number of windows.
func (wfa *WalkForwardAnalysis) WindowCount() int {
	return len(wfa.Windows)
}

// CompleteWindows returns the number of completed window results.
func (wfa *WalkForwardAnalysis) CompleteWindows() int {
	return len(wfa.Results)
}
