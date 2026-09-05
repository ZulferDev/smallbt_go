package runtime

import "errors"

var (
	// ErrInvalidMode is returned when execution mode is not recognized
	ErrInvalidMode = errors.New("invalid execution mode: must be backtest, paper, or live")
	
	// ErrMissingDataFeed is returned when data feed configuration is missing
	ErrMissingDataFeed = errors.New("data feed configuration is required")
	
	// ErrMissingBroker is returned when broker configuration is missing
	ErrMissingBroker = errors.New("broker configuration is required")
	
	// ErrMissingRiskLimits is returned when live mode is used without risk limits
	ErrMissingRiskLimits = errors.New("risk limits are required for live trading mode")
)
