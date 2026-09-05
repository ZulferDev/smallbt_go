package montecarlo

import (
	"fmt"
	"testing"
	"time"
)

func TestExampleRun(t *testing.T) {
	trades := []Trade{
		{
			ID:        1,
			EntryTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			ExitTime:  time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			NetPnL:    100.0,
			Return:    0.01,
		},
		{
			ID:        2,
			EntryTime: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			ExitTime:  time.Date(2024, 1, 2, 14, 0, 0, 0, time.UTC),
			NetPnL:    -50.0,
			Return:    -0.005,
		},
	}

	config := MCConfig{
		Simulations: 10,
		Seed:        42,
		Type:        TradeReshuffle,
	}

	runner := NewRunner(config, trades, 10000.0)
	result, err := runner.Run()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	fmt.Printf("Success! Ran %d simulations\n", len(result.Simulations))
	fmt.Printf("Mean Return: %.4f\n", result.Statistics.MeanReturn)
}
