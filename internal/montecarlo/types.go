package montecarlo

import "time"

// Trade represents a single trade for Monte Carlo analysis
type Trade struct {
	ID        int64
	EntryTime time.Time
	ExitTime  time.Time
	EntryPrice float64
	ExitPrice  float64
	Quantity  float64
	GrossPnL  float64
	Fees      float64
	NetPnL    float64
	Return    float64 // percentage
	MAE       float64 // Maximum Adverse Excursion
	MFE       float64 // Maximum Favorable Excursion
	Duration  time.Duration
}

// SimulationResult represents the result of a single Monte Carlo simulation
type SimulationResult struct {
	// Trade sequence
	Trades []Trade

	// Equity curve
	EquityCurve []float64 // equity at each timestamp

	// Metrics
	TotalReturn     float64
	TotalTrades     int
	WinningTrades   int
	LosingTrades    int
	WinRate         float64
	TotalPnL        float64
	MaxDrawdown     float64
	DrawdownFromPeak float64
	Sharpe          float64
}

// ConfidenceLevel represents percentile-based confidence intervals
type ConfidenceLevel struct {
	Percentile    float64 // e.g., 0.05, 0.25, 0.50, 0.75, 0.95
	TotalReturn   float64
	MaxDrawdown   float64
	WinRate       float64
	SharpeRatio   float64
}

// MCConfig configures Monte Carlo analysis parameters
type MCConfig struct {
	// Number of simulations to run
	Simulations int

	// Seed for deterministic randomization
	// If 0, uses time-based seed
	Seed int64

	// Analysis type
	Type MCAnalysisType

	// Optional: only reshuffle within these constraints
	// (e.g., preserve temporal structure at daily level)
	PreserveTemporalStructure bool

	// Optional: minimum trade duration to preserve
	MinTradeDuration time.Duration
}

// MCAnalysisType determines which reshuffling strategy to use
type MCAnalysisType int

const (
	// TradeReshuffle: randomly reorder trades (changes sequence, preserves outcomes)
	TradeReshuffle MCAnalysisType = iota

	// ReturnReshuffle: randomly reorder returns (changes PnL sequence)
	ReturnReshuffle

	// BootstrapReshuffle: randomly sample trades with replacement
	BootstrapReshuffle
)

// MCResult aggregates results from all simulations
type MCResult struct {
	// Configuration used
	Config MCConfig

	// All simulation results
	Simulations []SimulationResult

	// Aggregated statistics
	Statistics MCStatistics

	// Confidence intervals at key percentiles
	ConfidenceIntervals []ConfidenceLevel
}

// MCStatistics aggregates metrics across all simulations
type MCStatistics struct {
	// Return distribution
	MeanReturn           float64
	StdDevReturn         float64
	MinReturn            float64
	MaxReturn            float64
	MedianReturn         float64
	P05Return            float64 // 5th percentile
	P95Return            float64 // 95th percentile

	// Drawdown distribution
	MeanMaxDrawdown      float64
	StdDevMaxDrawdown    float64
	MinMaxDrawdown       float64
	MaxMaxDrawdown       float64
	MedianMaxDrawdown    float64
	P95MaxDrawdown       float64 // worst case (95th percentile)

	// Win rate distribution
	MeanWinRate          float64
	StdDevWinRate        float64
	MinWinRate           float64
	MaxWinRate           float64

	// Sharpe ratio distribution
	MeanSharpe           float64
	StdDevSharpe         float64
	MinSharpe            float64
	MaxSharpe            float64

	// Probability metrics
	ProbabilityOfRuin    float64 // % of simulations with total loss
	NegativeReturnCount  int     // number of simulations with negative return
	NegativeReturnRatio  float64
}

// DrawdownPoint represents a point in drawdown analysis
type DrawdownPoint struct {
	Timestamp   time.Time
	Equity      float64
	PeakEquity  float64
	Drawdown    float64
	DrawdownPct float64
}
