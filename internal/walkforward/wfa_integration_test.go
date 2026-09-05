package walkforward

import (
	"testing"
	"time"
	
	"github.com/ZulferDev/smallbt_go/internal/backtest"
	"github.com/ZulferDev/smallbt_go/internal/analytics"
)

func TestWalkForwardIntegration(t *testing.T) {
	// Create WFA with 3 windows
	// With Train=100, Test=50, Step=50, Total=300:
	// Window 0: bars 0-149 (train 0-99, test 100-149)
	// Window 1: bars 50-199 (train 50-149, test 150-199)
	// Window 2: bars 100-249 (train 100-199, test 200-249)
	// Window 3: bars 150-299 (train 150-249, test 250-299)
	// So Total=300 gives 4 windows
	config := WindowConfig{
		TrainBars: 100,
		TestBars:  50,
		StepBars:  50,
	}
	
	wfa, err := New(config)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	
	// Generate windows - with Step=50 and Train+Test=150, each window advances by 50
	// For Total=300 bars, we get 4 complete windows
	err = wfa.GenerateWindows(300)
	if err != nil {
		t.Fatalf("GenerateWindows() failed: %v", err)
	}
	
	if wfa.WindowCount() != 4 {
		t.Errorf("WindowCount: got %d, want 4", wfa.WindowCount())
	}
	
	// Add results for each window
	for i := 0; i < wfa.WindowCount(); i++ {
		equity := []backtest.EquityPoint{
			{Timestamp: time.Now(), Equity: 10000, Cash: 5000, Drawdown: 0, Exposure: 0.5},
			{Timestamp: time.Now().Add(time.Hour), Equity: 10500, Cash: 5250, Drawdown: 0, Exposure: 0.5},
		}
		
		trainMetrics := &analytics.Metrics{
			TotalReturn:  10.0 + float64(i),
			SharpeRatio:  1.5 + float64(i)*0.1,
			MaxDrawdown:  5.0,
			WinRate:      0.55,
			ProfitFactor: 1.8,
		}
		
		testMetrics := &analytics.Metrics{
			TotalReturn:  8.0 + float64(i),
			SharpeRatio:  1.2 + float64(i)*0.1,
			MaxDrawdown:  6.0,
			WinRate:      0.50,
			ProfitFactor: 1.5,
		}
		
		trainResult := &backtest.BacktestResult{
			TotalTrades: 15 + i,
			EquityCurve: equity,
			Metrics:     trainMetrics,
		}
		
		testResult := &backtest.BacktestResult{
			TotalTrades: 10 + i,
			EquityCurve: equity,
			Metrics:     testMetrics,
		}
		
		result := &WFWindowResult{
			WindowID:    i,
			TrainResult: trainResult,
			TestResult:  testResult,
		}
		
		err = wfa.AddWindowResult(i, result)
		if err != nil {
			t.Fatalf("AddWindowResult(%d) failed: %v", i, err)
		}
	}
	
	if wfa.CompleteWindows() != 4 {
		t.Errorf("CompleteWindows: got %d, want 4", wfa.CompleteWindows())
	}
	
	// Compute aggregate
	agg, err := wfa.ComputeAggregate()
	if err != nil {
		t.Fatalf("ComputeAggregate() failed: %v", err)
	}
	
	if agg == nil {
		t.Fatal("ComputeAggregate() returned nil")
	}
	
	// Verify aggregate values
	if agg.WindowCount != 4 {
		t.Errorf("WindowCount: got %d, want 4", agg.WindowCount)
	}
	
	// OOS trades should be sum of test trades: 10+11+12+13 = 46
	expectedTrades := 10 + 11 + 12 + 13
	if agg.TotalTrades != expectedTrades {
		t.Errorf("TotalTrades: got %d, want %d", agg.TotalTrades, expectedTrades)
	}
	
	// Total return should be average of test returns: (8+9+10+11)/4 = 9.5
	expectedReturn := (8.0 + 9.0 + 10.0 + 11.0) / 4.0
	if agg.TotalReturn != expectedReturn {
		t.Errorf("TotalReturn: got %f, want %f", agg.TotalReturn, expectedReturn)
	}
	
	// Sharpe ratio should be average of test Sharpe: (1.2+1.3+1.4+1.5)/4 = 1.35
	expectedSharpe := (1.2 + 1.3 + 1.4 + 1.5) / 4.0
	if agg.SharpeRatio != expectedSharpe {
		t.Errorf("SharpeRatio: got %f, want %f", agg.SharpeRatio, expectedSharpe)
	}
	
	// Generate report
	report := wfa.Report()
	if len(report) == 0 {
		t.Fatal("Report() returned empty string")
	}
	
	// Export to JSON
	jsonData, err := wfa.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON() failed: %v", err)
	}
	
	if len(jsonData) == 0 {
		t.Fatal("ExportToJSON() returned empty JSON")
	}
	
	// Export to CSV
	csvData, err := wfa.ExportToCSV()
	if err != nil {
		t.Fatalf("ExportToCSV() failed: %v", err)
	}
	
	if len(csvData) == 0 {
		t.Fatal("ExportToCSV() returned empty CSV")
	}
	
	t.Log("✅ Walk Forward Analysis integration test passed")
	t.Logf("📊 Total Trades (OOS): %d", agg.TotalTrades)
	t.Logf("📊 Total Return (OOS): %.2f%%", agg.TotalReturn)
	t.Logf("📊 Sharpe Ratio (OOS): %.2f", agg.SharpeRatio)
	t.Logf("📊 In-Sample Avg Sharpe: %.2f", agg.InSampleAvgSharpe)
	t.Logf("📊 Out-of-Sample Avg Sharpe: %.2f", agg.OutOfSampleAvgSharpe)
	t.Logf("📊 Sharpe Degradation: %.2f%%", agg.SharpeRatioDegradation*100)
}
