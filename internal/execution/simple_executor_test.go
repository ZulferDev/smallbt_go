package execution

import (
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleExecutorMarketOrder(t *testing.T) {
	config := Config{
		SlippageType:  "percentage",
		SlippageValue: 0.001,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
	}

	executor := NewSimpleExecutor(config)

	candle := &market.Candle{
		Open:   50000.0,
		High:   51000.0,
		Low:    49000.0,
		Close:  50500.0,
		Volume: 100.0,
	}

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	fill, err := executor.SimulateFill(req, candle)
	assert.NoError(t, err)
	assert.NotNil(t, fill)
	assert.Equal(t, order.OrderSideBuy, fill.Side)
	assert.Greater(t, fill.Price, 0.0)
	assert.Greater(t, fill.Fees, 0.0)
}

func TestSimpleExecutorLimitOrderFill(t *testing.T) {
	config := Config{
		SlippageType: "none",
		FeeMaker:     0.0002,
		FeeTaker:     0.0005,
	}

	executor := NewSimpleExecutor(config)

	candle := &market.Candle{
		Open:   50000.0,
		High:   51000.0,
		Low:    49000.0,
		Close:  50500.0,
		Volume: 100.0,
	}

	// Limit buy at 49500 (within range)
	price := 49500.0
	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeLimit,
		Quantity: 1.0,
		Price:    &price,
	}

	fill, err := executor.SimulateFill(req, candle)
	assert.NoError(t, err)
	assert.NotNil(t, fill)
	assert.Equal(t, 49500.0, fill.Price)
}

func TestSimpleExecutorLimitOrderNoFill(t *testing.T) {
	config := Config{
		SlippageType: "none",
		FeeMaker:     0.0002,
	}

	executor := NewSimpleExecutor(config)

	candle := &market.Candle{
		Open:   50000.0,
		High:   51000.0,
		Low:    49000.0,
		Close:  50500.0,
		Volume: 100.0,
	}

	// Limit buy at 48000 (below low)
	price := 48000.0
	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeLimit,
		Quantity: 1.0,
		Price:    &price,
	}

	fill, err := executor.SimulateFill(req, candle)
	assert.Error(t, err) // Should return error when not filled
	assert.Nil(t, fill)  // Should not fill
}

func TestSimpleExecutorStopOrder(t *testing.T) {
	config := Config{
		SlippageType: "none",
		FeeTaker:     0.0005,
	}

	executor := NewSimpleExecutor(config)

	candle := &market.Candle{
		Open:   50000.0,
		High:   51000.0,
		Low:    49000.0,
		Close:  50500.0,
		Volume: 100.0,
	}

	// Stop buy at 50800 (within range)
	stopPrice := 50800.0
	req := order.OrderRequest{
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      order.OrderSideBuy,
		Type:      order.OrderTypeStop,
		Quantity:  1.0,
		StopPrice: &stopPrice,
	}

	fill, err := executor.SimulateFill(req, candle)
	assert.NoError(t, err)
	assert.NotNil(t, fill)
	assert.GreaterOrEqual(t, fill.Price, 50800.0)
}

func TestSimpleExecutorSlippage(t *testing.T) {
	config := Config{
		SlippageType:  "percentage",
		SlippageValue: 0.01, // 1% slippage
		FeeTaker:      0.0005,
	}

	executor := NewSimpleExecutor(config)

	candle := &market.Candle{
		Open:   50000.0,
		High:   51000.0,
		Low:    49000.0,
		Close:  50500.0,
		Volume: 100.0,
	}

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	fill, err := executor.SimulateFill(req, candle)
	assert.NoError(t, err)
	assert.NotNil(t, fill)
	// With slippage, buy price should be higher than candle close
	assert.Greater(t, fill.Price, candle.Close)
}

func TestSimpleExecutorFees(t *testing.T) {
	config := Config{
		SlippageType: "none",
		FeeMaker:     0.0002,
		FeeTaker:     0.0005,
	}

	executor := NewSimpleExecutor(config)

	candle := &market.Candle{
		Open:   50000.0,
		High:   51000.0,
		Low:    49000.0,
		Close:  50500.0,
		Volume: 100.0,
	}

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	fill, err := executor.SimulateFill(req, candle)
	assert.NoError(t, err)
	assert.NotNil(t, fill)

	expectedFee := fill.Price * fill.Quantity * config.FeeTaker
	assert.InDelta(t, expectedFee, fill.Fees, 0.01)
}

func TestSimpleExecutorSellOrder(t *testing.T) {
	config := Config{
		SlippageType: "none",
		FeeTaker:     0.0005,
	}

	executor := NewSimpleExecutor(config)

	candle := &market.Candle{
		Open:   50000.0,
		High:   51000.0,
		Low:    49000.0,
		Close:  50500.0,
		Volume: 100.0,
	}

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideSell,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	fill, err := executor.SimulateFill(req, candle)
	assert.NoError(t, err)
	assert.NotNil(t, fill)
	assert.Equal(t, order.OrderSideSell, fill.Side)
}
