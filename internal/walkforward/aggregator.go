package walkforward

import (
	"math"

	"github.com/1jehuang/backtest/internal/analytics"
)

// ComputeAggregate computes aggregate metrics across all windows.
func (wfa *WalkForwardAnalysis) ComputeAggregate() (*WFAggregateResult, error) {
	if len(wfa.Results) == 0 {
		return nil, ErrNoResults
	}

	if len(wfa.Results) != len(wfa.Windows) {
		return nil, ErrIncompleteWindows
	}

	agg := &WFAggregateResult{
		WindowCount: len(wfa.Windows),
	}

	// Aggregate equity curve and metrics from all test (out-of-sample) windows
	var allEquity []analytics.EquityPoint
	var totalTrades int
	var totalReturn float64
	var inSampleSharpe []float64
	var outOfSampleSharpe []float64

	for _, result := range wfa.Results {
		// Aggregate test (out-of-sample) results
		if result.TestResult != nil {
			totalTrades += result.TestResult.TotalTrades
			if result.TestResult.Metrics != nil {
				totalReturn += result.TestResult.Metrics.TotalReturn
			}

			// Convert backtest.EquityPoint to analytics.EquityPoint
			for _, ep := range result.TestResult.EquityCurve {
				allEquity = append(allEquity, analytics.EquityPoint{
					Timestamp: ep.Timestamp,
					Equity:    ep.Equity,
					Cash:      ep.Cash,
					Drawdown:  ep.Drawdown,
					Exposure:  ep.Exposure,
				})
			}
		}

		// Collect Sharpe ratios for degradation analysis
		if result.TrainResult != nil && result.TrainResult.Metrics != nil {
			inSampleSharpe = append(inSampleSharpe, result.TrainResult.Metrics.SharpeRatio)
		}
		if result.TestResult != nil && result.TestResult.Metrics != nil {
			outOfSampleSharpe = append(outOfSampleSharpe, result.TestResult.Metrics.SharpeRatio)
		}
	}

	// Aggregate metrics from test windows
	agg.TotalTrades = totalTrades

	// Calculate aggregate metrics from test results
	if len(wfa.Results) > 0 && wfa.Results[0].TestResult != nil && wfa.Results[0].TestResult.Metrics != nil {
		// Use metrics from the first window as template
		firstMetrics := wfa.Results[0].TestResult.Metrics

		// Aggregate values across all test windows
		var sumTotalReturn, sumCAGR, sumSharpe, sumSortino float64
		var sumMaxDD, sumCalmar, sumWinRate, sumProfitFactor float64
		var sumExpectancy, sumAvgWin, sumAvgLoss, sumAvgTrade float64
		windowCount := float64(len(wfa.Results))

		for _, result := range wfa.Results {
			if result.TestResult != nil && result.TestResult.Metrics != nil {
				m := result.TestResult.Metrics
				sumTotalReturn += m.TotalReturn
				sumCAGR += m.CAGR
				sumSharpe += m.SharpeRatio
				sumSortino += m.SortinoRatio
				sumMaxDD += m.MaxDrawdown
				sumCalmar += m.CalmarRatio
				sumWinRate += m.WinRate
				sumProfitFactor += m.ProfitFactor
				sumExpectancy += m.Expectancy
				sumAvgWin += m.AvgWin
				sumAvgLoss += m.AvgLoss
				sumAvgTrade += m.AvgTrade
			}
		}

		// Average the metrics
		agg.TotalReturn = sumTotalReturn / windowCount
		agg.CAGR = sumCAGR / windowCount
		agg.SharpeRatio = sumSharpe / windowCount
		agg.SortinoRatio = sumSortino / windowCount
		agg.MaxDrawdown = sumMaxDD / windowCount
		agg.CalmarRatio = sumCalmar / windowCount
		agg.WinRate = sumWinRate / windowCount
		agg.ProfitFactor = sumProfitFactor / windowCount
		agg.Expectancy = sumExpectancy / windowCount
		agg.AverageWin = sumAvgWin / windowCount
		agg.AverageLoss = sumAvgLoss / windowCount
		agg.AverageTradeReturn = sumAvgTrade / windowCount

		// Copy initial values from first window
		_ = firstMetrics
	}

	// Calculate Sharpe ratio degradation
	agg.InSampleAvgSharpe = average(inSampleSharpe)
	agg.OutOfSampleAvgSharpe = average(outOfSampleSharpe)

	if agg.InSampleAvgSharpe != 0 {
		agg.SharpeRatioDegradation = (agg.InSampleAvgSharpe - agg.OutOfSampleAvgSharpe) / agg.InSampleAvgSharpe
	}

	wfa.AggregateResult = agg
	return agg, nil
}

// average calculates the average of a slice of floats.
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			sum += v
		}
	}

	return sum / float64(len(values))
}
