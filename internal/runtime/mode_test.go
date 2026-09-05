package runtime

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr error
	}{
		{
			name: "valid backtest config",
			config: &Config{
				Mode: ModeBacktest,
				DataFeed: DataFeedConfig{
					Type: "csv",
					Params: map[string]string{
						"path": "data.csv",
					},
				},
				Broker: BrokerConfig{
					Type: "simulated",
					Params: map[string]string{
						"initial_balance": "10000",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid paper config",
			config: &Config{
				Mode: ModePaper,
				DataFeed: DataFeedConfig{
					Type: "websocket",
					Params: map[string]string{
						"exchange": "binance",
					},
				},
				Broker: BrokerConfig{
					Type: "paper",
				},
			},
			wantErr: nil,
		},
		{
			name: "valid live config with risk limits",
			config: &Config{
				Mode: ModeLive,
				DataFeed: DataFeedConfig{
					Type: "websocket",
				},
				Broker: BrokerConfig{
					Type: "live",
				},
				RiskLimits: &RiskLimits{
					MaxPositionSize:  1000,
					MaxDailyLoss:     500,
					MaxOpenPositions: 3,
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid mode",
			config: &Config{
				Mode: "invalid",
				DataFeed: DataFeedConfig{
					Type: "csv",
				},
				Broker: BrokerConfig{
					Type: "simulated",
				},
			},
			wantErr: ErrInvalidMode,
		},
		{
			name: "missing data feed",
			config: &Config{
				Mode: ModeBacktest,
				Broker: BrokerConfig{
					Type: "simulated",
				},
			},
			wantErr: ErrMissingDataFeed,
		},
		{
			name: "missing broker",
			config: &Config{
				Mode: ModeBacktest,
				DataFeed: DataFeedConfig{
					Type: "csv",
				},
			},
			wantErr: ErrMissingBroker,
		},
		{
			name: "live mode without risk limits",
			config: &Config{
				Mode: ModeLive,
				DataFeed: DataFeedConfig{
					Type: "websocket",
				},
				Broker: BrokerConfig{
					Type: "live",
				},
			},
			wantErr: ErrMissingRiskLimits,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ModeChecks(t *testing.T) {
	tests := []struct {
		name         string
		mode         ExecutionMode
		isBacktest   bool
		isPaper      bool
		isLive       bool
		isRealTime   bool
	}{
		{
			name:         "backtest mode",
			mode:         ModeBacktest,
			isBacktest:   true,
			isPaper:      false,
			isLive:       false,
			isRealTime:   false,
		},
		{
			name:         "paper mode",
			mode:         ModePaper,
			isBacktest:   false,
			isPaper:      true,
			isLive:       false,
			isRealTime:   true,
		},
		{
			name:         "live mode",
			mode:         ModeLive,
			isBacktest:   false,
			isPaper:      false,
			isLive:       true,
			isRealTime:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Mode: tt.mode}

			if got := config.IsBacktest(); got != tt.isBacktest {
				t.Errorf("IsBacktest() = %v, want %v", got, tt.isBacktest)
			}
			if got := config.IsPaper(); got != tt.isPaper {
				t.Errorf("IsPaper() = %v, want %v", got, tt.isPaper)
			}
			if got := config.IsLive(); got != tt.isLive {
				t.Errorf("IsLive() = %v, want %v", got, tt.isLive)
			}
			if got := config.IsRealTime(); got != tt.isRealTime {
				t.Errorf("IsRealTime() = %v, want %v", got, tt.isRealTime)
			}
		})
	}
}
