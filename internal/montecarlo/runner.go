package montecarlo

import (
	"fmt"
)

// Runner executes Monte Carlo simulations
type Runner struct {
	config       MCConfig
	reshuffler   *Reshuffler
	initialTrades []Trade
	initialCapital float64
}

// NewRunner creates a new Monte Carlo runner
func NewRunner(config MCConfig, initialTrades []Trade, initialCapital float64) *Runner {
	if config.Simulations <= 0 {
		config.Simulations = 1000 // default
	}

	return &Runner{
		config:       config,
		reshuffler:   NewReshuffler(config.Seed),
		initialTrades: initialTrades,
		initialCapital: initialCapital,
	}
}

// Run executes all Monte Carlo simulations
func (r *Runner) Run() (*MCResult, error) {
	if len(r.initialTrades) == 0 {
		return nil, fmt.Errorf("no trades to analyze")
	}

	// Generate simulation results
	simulations := make([]SimulationResult, 0, r.config.Simulations)
	for i := 0; i < r.config.Simulations; i++ {
		sim, err := r.runSingleSimulation(i)
		if err != nil {
			return nil, fmt.Errorf("simulation %d failed: %w", i, err)
		}
		simulations = append(simulations, sim)
	}

	// Calculate aggregate statistics
	stats := CalculateStatistics(simulations)

	// Calculate confidence intervals
	confIntervals := r.calculateConfidenceIntervals(simulations)

	result := &MCResult{
		Config:              r.config,
		Simulations:         simulations,
		Statistics:          stats,
		ConfidenceIntervals: confIntervals,
	}

	return result, nil
}

// runSingleSimulation executes a single Monte Carlo simulation
func (r *Runner) runSingleSimulation(simID int) (SimulationResult, error) {
	var shuffledTrades []Trade

	switch r.config.Type {
	case TradeReshuffle:
		shuffledTrades = r.reshuffler.ShuffleTrades(r.initialTrades)
	case ReturnReshuffle:
		shuffledTrades = r.reshuffler.ShuffleReturns(r.initialTrades)
	case BootstrapReshuffle:
		shuffledTrades = r.reshuffler.BootstrapTrades(r.initialTrades, len(r.initialTrades))
	default:
		shuffledTrades = r.reshuffler.ShuffleTrades(r.initialTrades)
	}

	// Build equity curve
	equityCurve := BuildEquityCurve(r.initialCapital, shuffledTrades)

	// Calculate metrics
	totalReturn := calculateTotalReturn(equityCurve)
	maxDrawdown := calculateDrawdown(equityCurve)
	winRate := CalculateWinRate(shuffledTrades)
	// Simplified Sharpe calculation (annualized, 0 risk-free rate for now)
	sharpe := CalculateSharpe(shuffledTrades, 0)

	// Count trades by type
	winningTrades := 0
	losingTrades := 0
	totalPnL := 0.0
	for _, trade := range shuffledTrades {
		totalPnL += trade.NetPnL
		if trade.NetPnL > 0 {
			winningTrades++
		} else if trade.NetPnL < 0 {
			losingTrades++
		}
	}

	return SimulationResult{
		Trades:           shuffledTrades,
		EquityCurve:      equityCurve,
		TotalReturn:      totalReturn,
		TotalTrades:      len(shuffledTrades),
		WinningTrades:    winningTrades,
		LosingTrades:     losingTrades,
		WinRate:          winRate,
		TotalPnL:         totalPnL,
		MaxDrawdown:      maxDrawdown,
		DrawdownFromPeak: maxDrawdown,
		Sharpe:           sharpe,
	}, nil
}

// calculateTotalReturn calculates total return from equity curve
func calculateTotalReturn(equityCurve []float64) float64 {
	if len(equityCurve) < 2 {
		return 0
	}
	return (equityCurve[len(equityCurve)-1] - equityCurve[0]) / equityCurve[0]
}

// calculateConfidenceIntervals computes confidence intervals at key percentiles
func (r *Runner) calculateConfidenceIntervals(simulations []SimulationResult) []ConfidenceLevel {
	percentiles := []float64{0.05, 0.25, 0.50, 0.75, 0.95}

	// Extract metrics for each simulation
	returns := make([]float64, len(simulations))
	drawdowns := make([]float64, len(simulations))
	winRates := make([]float64, len(simulations))
	sharpes := make([]float64, len(simulations))

	for i, sim := range simulations {
		returns[i] = sim.TotalReturn
		drawdowns[i] = sim.MaxDrawdown
		winRates[i] = sim.WinRate
		sharpes[i] = sim.Sharpe
	}

	// Calculate confidence intervals
	intervals := make([]ConfidenceLevel, 0, len(percentiles))
	for _, p := range percentiles {
		interval := ConfidenceLevel{
			Percentile:    p,
			TotalReturn:   CalculatePercentile(returns, p),
			MaxDrawdown:   CalculatePercentile(drawdowns, p),
			WinRate:       CalculatePercentile(winRates, p),
			SharpeRatio:   CalculatePercentile(sharpes, p),
		}
		intervals = append(intervals, interval)
	}

	return intervals
}
