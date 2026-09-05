package runtime

// ExecutionMode defines how the strategy runs
type ExecutionMode string

const (
	// ModeBacktest runs strategy against historical data with simulated execution
	ModeBacktest ExecutionMode = "backtest"
	
	// ModePaper runs strategy with real-time data but simulated execution (no real money)
	ModePaper ExecutionMode = "paper"
	
	// ModeLive runs strategy with real-time data and real exchange execution
	ModeLive ExecutionMode = "live"
)

// Config holds runtime configuration for strategy execution
type Config struct {
	// Mode determines execution behavior (backtest, paper, or live)
	Mode ExecutionMode
	
	// DataFeed configuration
	DataFeed DataFeedConfig
	
	// Broker configuration
	Broker BrokerConfig
	
	// RiskLimits for live trading (optional, recommended for live mode)
	RiskLimits *RiskLimits
}

// DataFeedConfig specifies the data source
type DataFeedConfig struct {
	// Type of data feed: "csv", "parquet", "websocket", "rest"
	Type string
	
	// Params holds type-specific configuration
	// CSV: {"path": "data/BTCUSDT.csv"}
	// WebSocket: {"exchange": "binance", "symbol": "BTCUSDT"}
	Params map[string]string
}

// BrokerConfig specifies the order execution backend
type BrokerConfig struct {
	// Type of broker: "simulated", "paper", "live"
	Type string
	
	// Params holds type-specific configuration
	// Simulated: {"initial_balance": "10000", "fee": "0.001"}
	// Live: {"exchange": "binance", "api_key_env": "BINANCE_API_KEY"}
	Params map[string]string
}

// RiskLimits define safety constraints for live trading
type RiskLimits struct {
	// MaxPositionSize in quote currency (e.g., max $1000 per position)
	MaxPositionSize float64
	
	// MaxDailyLoss in quote currency (kill switch threshold)
	MaxDailyLoss float64
	
	// RequireConfirm forces manual confirmation before each order (live mode)
	RequireConfirm bool
	
	// MaxOpenPositions limits concurrent positions
	MaxOpenPositions int
	
	// MaxLeverage caps leverage (for futures/margin)
	MaxLeverage float64
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Mode != ModeBacktest && c.Mode != ModePaper && c.Mode != ModeLive {
		return ErrInvalidMode
	}
	
	if c.DataFeed.Type == "" {
		return ErrMissingDataFeed
	}
	
	if c.Broker.Type == "" {
		return ErrMissingBroker
	}
	
	// Live mode must have risk limits
	if c.Mode == ModeLive && c.RiskLimits == nil {
		return ErrMissingRiskLimits
	}
	
	return nil
}

// IsBacktest returns true if running in backtest mode
func (c *Config) IsBacktest() bool {
	return c.Mode == ModeBacktest
}

// IsPaper returns true if running in paper trading mode
func (c *Config) IsPaper() bool {
	return c.Mode == ModePaper
}

// IsLive returns true if running in live trading mode
func (c *Config) IsLive() bool {
	return c.Mode == ModeLive
}

// IsRealTime returns true if using real-time data (paper or live)
func (c *Config) IsRealTime() bool {
	return c.Mode == ModePaper || c.Mode == ModeLive
}
