package analytics

import (
	"math"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// DefaultAnalyzer implements the Analyzer interface.
type DefaultAnalyzer struct{}

// NewAnalyzer creates a new DefaultAnalyzer.
func NewAnalyzer() Analyzer {
	return &DefaultAnalyzer{}
}

// Analyze calculates all performance metrics.
func (a *DefaultAnalyzer) Analyze(input AnalysisInput) *Metrics {
	metrics := &Metrics{}

	if len(input.TradeHistory) == 0 {
		return metrics
	}

	// Calculate returns
	metrics.TotalReturn = calculateTotalReturn(input.InitialCash, input.FinalEquity)
	metrics.CAGR = calculateCAGR(metrics.TotalReturn, input.StartTime, input.EndTime)

	// Calculate drawdown metrics
	metrics.MaxDrawdown, metrics.MaxDrawdownDate, metrics.AvgDrawdown = calculateDrawdowns(input.EquityCurve)

	// Calculate trade statistics
	metrics.TotalTrades = len(input.TradeHistory)
	metrics.WinningTrades, metrics.LosingTrades = countTrades(input.TradeHistory)
	metrics.WinRate = float64(metrics.WinningTrades) / float64(metrics.TotalTrades)

	// Calculate PnL statistics
	metrics.GrossProfit, metrics.GrossLoss, metrics.NetProfit = calculatePnL(input.TradeHistory)
	metrics.TotalFees = calculateTotalFees(input.TradeHistory)
	metrics.ProfitFactor = calculateProfitFactor(metrics.GrossProfit, metrics.GrossLoss)

	// Calculate averages
	metrics.AvgTrade = metrics.NetProfit / float64(metrics.TotalTrades)
	if metrics.WinningTrades > 0 {
		metrics.AvgWin = metrics.GrossProfit / float64(metrics.WinningTrades)
	}
	if metrics.LosingTrades > 0 {
		metrics.AvgLoss = metrics.GrossLoss / float64(metrics.LosingTrades)
	}

	// Calculate best/worst
	metrics.LargestWin, metrics.LargestLoss = findExtremes(input.TradeHistory)

	// Calculate expectancy
	metrics.Expectancy = metrics.AvgTrade

	// Calculate exposure
	metrics.AvgExposure = calculateAverageExposure(input.EquityCurve)

	// Calculate risk-adjusted returns
	returns := calculateReturns(input.EquityCurve)
	metrics.SharpeRatio = calculateSharpeRatio(returns, input.RiskFreeRate)
	metrics.SortinoRatio = calculateSortinoRatio(returns, input.RiskFreeRate)
	metrics.CalmarRatio = calculateCalmarRatio(metrics.TotalReturn, metrics.MaxDrawdown)

	return metrics
}

// calculateTotalReturn calculates the total return percentage.
func calculateTotalReturn(initialCash, finalEquity float64) float64 {
	if initialCash <= 0 {
		return 0
	}
	return (finalEquity - initialCash) / initialCash
}

// calculateCAGR calculates Compound Annual Growth Rate.
func calculateCAGR(totalReturn float64, startTime, endTime time.Time) float64 {
	days := endTime.Sub(startTime).Hours() / 24
	if days <= 0 {
		return 0
	}

	years := days / 365.25
	if years <= 0 {
		return 0
	}

	return math.Pow(1+totalReturn, 1/years) - 1
}

// calculateDrawdowns calculates max, max date, and average drawdown.
func calculateDrawdowns(equityCurve []EquityPoint) (float64, time.Time, float64) {
	if len(equityCurve) == 0 {
		return 0, time.Time{}, 0
	}

	maxDD := 0.0
	maxDDDate := time.Time{}
	var totalDrawdown float64
	count := 0

	for _, point := range equityCurve {
		if point.Drawdown < 0 {
			if point.Drawdown < maxDD {
				maxDD = point.Drawdown
				maxDDDate = point.Timestamp
			}
			totalDrawdown += point.Drawdown
			count++
		}
	}

	avgDD := 0.0
	if count > 0 {
		avgDD = totalDrawdown / float64(count)
	}

	return maxDD, maxDDDate, avgDD
}

// countTrades counts winning and losing trades.
func countTrades(trades []portfolio.Trade) (int, int) {
	wins, losses := 0, 0
	for _, trade := range trades {
		if trade.NetPnL > 0 {
			wins++
		} else {
			losses++
		}
	}
	return wins, losses
}

// calculatePnL calculates gross profit, gross loss, and net profit.
func calculatePnL(trades []portfolio.Trade) (float64, float64, float64) {
	var grossProfit, grossLoss, netProfit float64
	for _, trade := range trades {
		netProfit += trade.NetPnL
		if trade.NetPnL > 0 {
			grossProfit += trade.NetPnL
		} else {
			grossLoss += trade.NetPnL
		}
	}
	return grossProfit, grossLoss, netProfit
}

// calculateTotalFees calculates total fees paid.
func calculateTotalFees(trades []portfolio.Trade) float64 {
	var total float64
	for _, trade := range trades {
		total += trade.Fees
	}
	return total
}

// calculateProfitFactor calculates profit factor (gross profit / |gross loss|).
func calculateProfitFactor(grossProfit, grossLoss float64) float64 {
	if grossLoss == 0 {
		if grossProfit > 0 {
			return grossProfit
		}
		return 0
	}
	return grossProfit / math.Abs(grossLoss)
}

// findExtremes finds the largest win and loss.
func findExtremes(trades []portfolio.Trade) (float64, float64) {
	var largestWin, largestLoss float64
	for _, trade := range trades {
		if trade.NetPnL > largestWin {
			largestWin = trade.NetPnL
		}
		if trade.NetPnL < largestLoss {
			largestLoss = trade.NetPnL
		}
	}
	return largestWin, largestLoss
}

// calculateAverageExposure calculates average exposure.
func calculateAverageExposure(equityCurve []EquityPoint) float64 {
	if len(equityCurve) == 0 {
		return 0
	}

	var total float64
	for _, point := range equityCurve {
		total += point.Exposure
	}
	return total / float64(len(equityCurve))
}

// calculateReturns calculates period-over-period returns from equity curve.
func calculateReturns(equityCurve []EquityPoint) []float64 {
	if len(equityCurve) < 2 {
		return nil
	}

	returns := make([]float64, 0, len(equityCurve)-1)
	for i := 1; i < len(equityCurve); i++ {
		prev := equityCurve[i-1].Equity
		curr := equityCurve[i].Equity
		if prev > 0 {
			returns = append(returns, (curr-prev)/prev)
		}
	}
	return returns
}

// calculateSharpeRatio calculates the Sharpe ratio.
// Assumes 252 trading days per year for crypto/equity markets.
func calculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	mean := calculateMean(returns)
	stdDev := calculateStdDev(returns, mean)

	if stdDev == 0 {
		return 0
	}

	// Annual risk-free rate to per-period risk-free rate
	periodRiskFreeRate := riskFreeRate / 252

	// Sharpe ratio = (mean return - risk-free rate) / std dev * sqrt(252)
	return (mean - periodRiskFreeRate) / stdDev * math.Sqrt(252)
}

// calculateSortinoRatio calculates the Sortino ratio.
// Only considers downside volatility (negative returns).
func calculateSortinoRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	mean := calculateMean(returns)
	downSideDeviation := calculateDownsideDeviation(returns, mean)

	if downSideDeviation == 0 {
		return 0
	}

	// Annual risk-free rate to per-period risk-free rate
	periodRiskFreeRate := riskFreeRate / 252

	// Sortino ratio = (mean return - risk-free rate) / downside std dev * sqrt(252)
	return (mean - periodRiskFreeRate) / downSideDeviation * math.Sqrt(252)
}

// calculateCalmarRatio calculates the Calmar ratio.
// Calmar = Annual Return / Max Drawdown
func calculateCalmarRatio(totalReturn, maxDrawdown float64) float64 {
	if maxDrawdown == 0 || maxDrawdown > 0 {
		return 0
	}
	return totalReturn / math.Abs(maxDrawdown)
}

// calculateMean calculates the mean of returns.
func calculateMean(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	return sum / float64(len(returns))
}

// calculateStdDev calculates the standard deviation of returns.
func calculateStdDev(returns []float64, mean float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	var variance float64
	for _, r := range returns {
		diff := r - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))

	return math.Sqrt(variance)
}

// calculateDownsideDeviation calculates downside deviation (only negative returns).
func calculateDownsideDeviation(returns []float64, mean float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	var variance float64
	count := 0

	for _, r := range returns {
		if r < 0 {
			diff := r - mean
			variance += diff * diff
			count++
		}
	}

	if count == 0 {
		return 0
	}

	variance /= float64(count)
	return math.Sqrt(variance)
}
