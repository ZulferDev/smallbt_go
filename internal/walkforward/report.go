package walkforward

import (
	"fmt"
)

// ExportToJSON exports aggregate results to JSON format.
func (wfa *WalkForwardAnalysis) ExportToJSON() (map[string]interface{}, error) {
	if wfa.AggregateResult == nil {
		return nil, ErrNoResults
	}

	return map[string]interface{}{
		"window_config": map[string]interface{}{
			"train_bars": wfa.Config.TrainBars,
			"test_bars":  wfa.Config.TestBars,
			"step_bars":  wfa.Config.StepBars,
		},
		"total_windows":     len(wfa.Windows),
		"completed_windows": len(wfa.Results),
		"aggregate": map[string]interface{}{
			"total_trades":             wfa.AggregateResult.TotalTrades,
			"total_return":             wfa.AggregateResult.TotalReturn,
			"cagr":                     wfa.AggregateResult.CAGR,
			"sharpe_ratio":             wfa.AggregateResult.SharpeRatio,
			"sortino_ratio":            wfa.AggregateResult.SortinoRatio,
			"max_drawdown":             wfa.AggregateResult.MaxDrawdown,
			"calmar_ratio":             wfa.AggregateResult.CalmarRatio,
			"win_rate":                 wfa.AggregateResult.WinRate,
			"profit_factor":            wfa.AggregateResult.ProfitFactor,
			"expectancy":               wfa.AggregateResult.Expectancy,
			"average_win":              wfa.AggregateResult.AverageWin,
			"average_loss":             wfa.AggregateResult.AverageLoss,
			"average_trade_return":     wfa.AggregateResult.AverageTradeReturn,
			"in_sample_avg_sharpe":     wfa.AggregateResult.InSampleAvgSharpe,
			"out_of_sample_avg_sharpe": wfa.AggregateResult.OutOfSampleAvgSharpe,
			"sharpe_ratio_degradation": wfa.AggregateResult.SharpeRatioDegradation,
		},
		"windows": wfa.exportWindows(),
	}, nil
}

func (wfa *WalkForwardAnalysis) exportWindows() []map[string]interface{} {
	var windows []map[string]interface{}

	for windowID := 0; windowID < len(wfa.Windows); windowID++ {
		window := wfa.Windows[windowID]
		result, exists := wfa.Results[windowID]

		windowData := map[string]interface{}{
			"window_id":   windowID,
			"train_start": window.TrainStart,
			"train_end":   window.TrainEnd,
			"test_start":  window.TestStart,
			"test_end":    window.TestEnd,
		}

		if exists {
			if result.TrainResult != nil && result.TrainResult.Metrics != nil {
				windowData["train_return"] = result.TrainResult.Metrics.TotalReturn
				windowData["train_sharpe"] = result.TrainResult.Metrics.SharpeRatio
			}
			if result.TestResult != nil && result.TestResult.Metrics != nil {
				windowData["test_return"] = result.TestResult.Metrics.TotalReturn
				windowData["test_sharpe"] = result.TestResult.Metrics.SharpeRatio
				windowData["test_trades"] = result.TestResult.TotalTrades
			}
		}

		windows = append(windows, windowData)
	}

	return windows
}

// ExportToCSV returns CSV data for all windows.
func (wfa *WalkForwardAnalysis) ExportToCSV() (string, error) {
	if wfa.AggregateResult == nil {
		return "", ErrNoResults
	}

	// Header
	csv := "window_id,train_start,train_end,test_start,test_end,train_return,test_return,train_sharpe,test_sharpe\n"

	// Data rows
	for windowID := 0; windowID < len(wfa.Windows); windowID++ {
		window := wfa.Windows[windowID]
		result, exists := wfa.Results[windowID]
		if !exists {
			continue
		}

		trainReturn := "N/A"
		testReturn := "N/A"
		trainSharpe := "N/A"
		testSharpe := "N/A"

		if result.TrainResult != nil && result.TrainResult.Metrics != nil {
			trainReturn = fmt.Sprintf("%.2f", result.TrainResult.Metrics.TotalReturn)
			trainSharpe = fmt.Sprintf("%.2f", result.TrainResult.Metrics.SharpeRatio)
		}
		if result.TestResult != nil && result.TestResult.Metrics != nil {
			testReturn = fmt.Sprintf("%.2f", result.TestResult.Metrics.TotalReturn)
			testSharpe = fmt.Sprintf("%.2f", result.TestResult.Metrics.SharpeRatio)
		}

		csv += fmt.Sprintf("%d,%d,%d,%d,%d,%s,%s,%s,%s\n",
			windowID,
			window.TrainStart,
			window.TrainEnd,
			window.TestStart,
			window.TestEnd,
			trainReturn,
			testReturn,
			trainSharpe,
			testSharpe)
	}

	return csv, nil
}

// Report generates a human-readable report.
func (wfa *WalkForwardAnalysis) Report() string {
	if wfa.AggregateResult == nil {
		return "Walk Forward Analysis: No results available"
	}

	report := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	report += "WALK FORWARD ANALYSIS REPORT\n"
	report += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"

	report += "WINDOW CONFIGURATION\n"
	report += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	report += fmt.Sprintf("Training Bars:  %d\n", wfa.Config.TrainBars)
	report += fmt.Sprintf("Testing Bars:   %d\n", wfa.Config.TestBars)
	report += fmt.Sprintf("Step Size:      %d\n\n", wfa.Config.StepBars)

	report += "OUT-OF-SAMPLE PERFORMANCE\n"
	report += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	report += fmt.Sprintf("Total Windows:  %d\n", len(wfa.Windows))
	report += fmt.Sprintf("Total Trades:   %d\n\n", wfa.AggregateResult.TotalTrades)

	report += fmt.Sprintf("Return:         %.2f%%\n", wfa.AggregateResult.TotalReturn)
	report += fmt.Sprintf("CAGR:           %.2f%%\n", wfa.AggregateResult.CAGR)
	report += fmt.Sprintf("Sharpe Ratio:   %.2f\n", wfa.AggregateResult.SharpeRatio)
	report += fmt.Sprintf("Sortino Ratio:  %.2f\n", wfa.AggregateResult.SortinoRatio)
	report += fmt.Sprintf("Max Drawdown:   %.2f%%\n", wfa.AggregateResult.MaxDrawdown)
	report += fmt.Sprintf("Profit Factor:  %.2f\n", wfa.AggregateResult.ProfitFactor)
	report += fmt.Sprintf("Win Rate:       %.2f%%\n", wfa.AggregateResult.WinRate)
	report += fmt.Sprintf("Expectancy:     %.2fR\n\n", wfa.AggregateResult.Expectancy)

	report += "IN-SAMPLE vs OUT-OF-SAMPLE\n"
	report += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	report += fmt.Sprintf("In-Sample Avg Sharpe:     %.2f\n", wfa.AggregateResult.InSampleAvgSharpe)
	report += fmt.Sprintf("Out-of-Sample Avg Sharpe: %.2f\n", wfa.AggregateResult.OutOfSampleAvgSharpe)
	report += fmt.Sprintf("Sharpe Degradation:       %.2f%%\n\n", wfa.AggregateResult.SharpeRatioDegradation*100)

	report += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	report += "WINDOW BREAKDOWN\n"
	report += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

	for windowID := 0; windowID < len(wfa.Windows); windowID++ {
		window := wfa.Windows[windowID]
		result, exists := wfa.Results[windowID]
		if !exists {
			continue
		}

		report += fmt.Sprintf("\nWindow %d:\n", windowID)

		trainReturn := "N/A"
		if result.TrainResult != nil && result.TrainResult.Metrics != nil {
			trainReturn = fmt.Sprintf("%.2f%%", result.TrainResult.Metrics.TotalReturn)
		}

		testReturn := "N/A"
		testSharpe := "N/A"
		if result.TestResult != nil && result.TestResult.Metrics != nil {
			testReturn = fmt.Sprintf("%.2f%%", result.TestResult.Metrics.TotalReturn)
			testSharpe = fmt.Sprintf("%.2f", result.TestResult.Metrics.SharpeRatio)
		}

		report += fmt.Sprintf("  Train: %d-%d (Return: %s)\n", window.TrainStart, window.TrainEnd, trainReturn)
		report += fmt.Sprintf("  Test:  %d-%d (Return: %s, Sharpe: %s)\n", window.TestStart, window.TestEnd, testReturn, testSharpe)
	}

	return report
}
