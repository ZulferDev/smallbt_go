package montecarlo

import (
	"math/rand"
	"sort"
	"time"
)

// Reshuffler performs trade/return reshuffling for Monte Carlo simulations
type Reshuffler struct {
	rng *rand.Rand
}

// NewReshuffler creates a new Reshuffler with the given seed
func NewReshuffler(seed int64) *Reshuffler {
	var rng *rand.Rand
	if seed == 0 {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		rng = rand.New(rand.NewSource(seed))
	}
	return &Reshuffler{rng: rng}
}

// ShuffleTrades randomly reorders trades using Fisher-Yates algorithm
func (r *Reshuffler) ShuffleTrades(trades []Trade) []Trade {
	shuffled := make([]Trade, len(trades))
	copy(shuffled, trades)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// ShuffleReturns randomly reorders returns (NetPnL values) while keeping other trade data
func (r *Reshuffler) ShuffleReturns(trades []Trade) []Trade {
	// Extract returns
	returns := make([]float64, len(trades))
	for i, trade := range trades {
		returns[i] = trade.NetPnL
	}

	// Shuffle returns
	for i := len(returns) - 1; i > 0; i-- {
		j := r.rng.Intn(i + 1)
		returns[i], returns[j] = returns[j], returns[i]
	}

	// Create new trades with shuffled returns
	shuffled := make([]Trade, len(trades))
	for i, trade := range trades {
		shuffled[i] = trade
		shuffled[i].NetPnL = returns[i]
		// Recalculate gross PnL (approximation)
		shuffled[i].GrossPnL = returns[i] + trade.Fees
	}

	return shuffled
}

// BootstrapTrades randomly samples trades with replacement
func (r *Reshuffler) BootstrapTrades(trades []Trade, sampleSize int) []Trade {
	if sampleSize <= 0 {
		sampleSize = len(trades)
	}

	sampled := make([]Trade, sampleSize)
	for i := 0; i < sampleSize; i++ {
		idx := r.rng.Intn(len(trades))
		sampled[i] = trades[idx]
	}

	return sampled
}

// calculateDrawdown calculates maximum drawdown from equity curve
func calculateDrawdown(equityCurve []float64) float64 {
	if len(equityCurve) == 0 {
		return 0
	}

	maxEquity := equityCurve[0]
	maxDrawdown := 0.0

	for _, equity := range equityCurve {
		if equity > maxEquity {
			maxEquity = equity
		}

		drawdown := (maxEquity - equity) / maxEquity
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// BuildEquityCurve constructs equity curve from initial capital and trades
func BuildEquityCurve(initialCapital float64, trades []Trade) []float64 {
	if len(trades) == 0 {
		return []float64{initialCapital}
	}

	equity := initialCapital
	curve := make([]float64, 0, len(trades)+1)
	curve = append(curve, equity)

	for _, trade := range trades {
		equity += trade.NetPnL
		curve = append(curve, equity)
	}

	return curve
}

// CalculateWinRate calculates win rate from trades
func CalculateWinRate(trades []Trade) float64 {
	if len(trades) == 0 {
		return 0
	}

	wins := 0
	for _, trade := range trades {
		if trade.NetPnL > 0 {
			wins++
		}
	}

	return float64(wins) / float64(len(trades))
}

// CalculateSharpe calculates Sharpe ratio from returns
func CalculateSharpe(trades []Trade, riskFreeRate float64) float64 {
	if len(trades) == 0 {
		return 0
	}

	// Calculate average return
	returns := make([]float64, len(trades))
	for i, trade := range trades {
		returns[i] = trade.NetPnL
	}

	// Mean return
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))

	// Standard deviation
	variance := 0.0
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	variance /= float64(len(returns))
	stdDev := 0.0
	if variance > 0 {
		stdDev = sqrt(variance)
	}

	if stdDev == 0 {
		return 0
	}

	// Annualized Sharpe (assuming daily returns)
	// This is simplified; real implementation would need to know timeframe
	return (mean - riskFreeRate) / stdDev
}

func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	// Newton's method for sqrt
	z := 1.0
	for i := 0; i < 100; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

// CalculatePercentile calculates the value at a given percentile
func CalculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Linear interpolation
	index := percentile * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// CalculateStatistics computes aggregate statistics from simulation results
func CalculateStatistics(results []SimulationResult) MCStatistics {
	if len(results) == 0 {
		return MCStatistics{}
	}

	stats := MCStatistics{}

	// Collect all values
	returns := make([]float64, len(results))
	drawdowns := make([]float64, len(results))
	winRates := make([]float64, len(results))
	sharpes := make([]float64, len(results))

	for i, sim := range results {
		returns[i] = sim.TotalReturn
		drawdowns[i] = sim.MaxDrawdown
		winRates[i] = sim.WinRate
		sharpes[i] = sim.Sharpe
	}

	// Calculate return statistics
	stats.MeanReturn = mean(returns)
	stats.StdDevReturn = stdDev(returns)
	stats.MinReturn = min(returns)
	stats.MaxReturn = max(returns)
	stats.MedianReturn = CalculatePercentile(returns, 0.50)
	stats.P05Return = CalculatePercentile(returns, 0.05)
	stats.P95Return = CalculatePercentile(returns, 0.95)

	// Calculate drawdown statistics
	stats.MeanMaxDrawdown = mean(drawdowns)
	stats.StdDevMaxDrawdown = stdDev(drawdowns)
	stats.MinMaxDrawdown = min(drawdowns)
	stats.MaxMaxDrawdown = max(drawdowns)
	stats.MedianMaxDrawdown = CalculatePercentile(drawdowns, 0.50)
	stats.P95MaxDrawdown = CalculatePercentile(drawdowns, 0.95)

	// Calculate win rate statistics
	stats.MeanWinRate = mean(winRates)
	stats.StdDevWinRate = stdDev(winRates)
	stats.MinWinRate = min(winRates)
	stats.MaxWinRate = max(winRates)

	// Calculate Sharpe statistics
	stats.MeanSharpe = mean(sharpes)
	stats.StdDevSharpe = stdDev(sharpes)
	stats.MinSharpe = min(sharpes)
	stats.MaxSharpe = max(sharpes)

	// Calculate probability of ruin (negative total return)
	negativeCount := 0
	for _, r := range returns {
		if r < 0 {
			negativeCount++
		}
	}
	stats.NegativeReturnCount = negativeCount
	stats.NegativeReturnRatio = float64(negativeCount) / float64(len(results))
	stats.ProbabilityOfRuin = stats.NegativeReturnRatio

	return stats
}

// Helper functions
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	avg := mean(values)
	variance := 0.0
	for _, v := range values {
		variance += (v - avg) * (v - avg)
	}
	variance /= float64(len(values))
	return sqrt(variance)
}

func min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values {
		if v < m {
			m = v
		}
	}
	return m
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values {
		if v > m {
			m = v
		}
	}
	return m
}
