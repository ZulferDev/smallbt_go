package montecarlo

import "fmt"

// String implements Stringer for MCAnalysisType
func (m MCAnalysisType) String() string {
	switch m {
	case TradeReshuffle:
		return "TradeReshuffle"
	case ReturnReshuffle:
		return "ReturnReshuffle"
	case BootstrapReshuffle:
		return "BootstrapReshuffle"
	default:
		return "Unknown"
	}
}

// Validate checks if MCConfig is valid
func (c MCConfig) Validate() error {
	if c.Simulations <= 0 {
		return fmt.Errorf("simulations must be positive, got %d", c.Simulations)
	}
	return nil
}
