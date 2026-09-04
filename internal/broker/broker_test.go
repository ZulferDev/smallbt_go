package broker

import (
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/execution"
	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/order"
	"github.com/stretchr/testify/assert"
)

func TestBrokerSubmitOrder(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	ord, err := broker.SubmitOrder(req, time.Now())
	assert.NoError(t, err)
	assert.NotNil(t, ord)
	assert.Equal(t, order.OrderStatusAccepted, ord.Status)
}

func TestBrokerProcessMarketOrder(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	ord, err := broker.SubmitOrder(req, time.Now())
	assert.NoError(t, err)

	candle := &market.Candle{
		Open:   50000.0,
		High:   50100.0,
		Low:    49900.0,
		Close:  50050.0,
		Volume: 100.0,
	}

	filled, err := broker.ProcessPendingOrders(candle, time.Now())
	assert.NoError(t, err)
	assert.Len(t, filled, 1)
	assert.Equal(t, ord.ID, filled[0])

	updated, _ := broker.GetOrder(ord.ID)
	assert.Equal(t, order.OrderStatusFilled, updated.Status)
	assert.Equal(t, 1.0, updated.FilledQty)
}

func TestBrokerLimitOrderNotFilled(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	price := 49000.0 // Below candle low
	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeLimit,
		Quantity: 1.0,
		Price:    &price,
	}

	_, err := broker.SubmitOrder(req, time.Now())
	assert.NoError(t, err)

	candle := &market.Candle{
		Open:   50000.0,
		High:   50100.0,
		Low:    49900.0,
		Close:  50050.0,
		Volume: 100.0,
	}

	filled, err := broker.ProcessPendingOrders(candle, time.Now())
	assert.NoError(t, err)
	assert.Len(t, filled, 0) // Limit order not hit

	pending := broker.GetPendingOrders()
	assert.Len(t, pending, 1)
}

func TestBrokerLimitOrderFilled(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	price := 49950.0 // Within candle range
	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeLimit,
		Quantity: 1.0,
		Price:    &price,
	}

	ord, err := broker.SubmitOrder(req, time.Now())
	assert.NoError(t, err)

	candle := &market.Candle{
		Open:   50000.0,
		High:   50100.0,
		Low:    49900.0,
		Close:  50050.0,
		Volume: 100.0,
	}

	filled, err := broker.ProcessPendingOrders(candle, time.Now())
	assert.NoError(t, err)
	assert.Len(t, filled, 1)

	updated, _ := broker.GetOrder(ord.ID)
	assert.Equal(t, order.OrderStatusFilled, updated.Status)
}

func TestBrokerStopOrderNotTriggered(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	stopPrice := 48000.0 // Below candle low
	req := order.OrderRequest{
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      order.OrderSideSell,
		Type:      order.OrderTypeStop,
		Quantity:  1.0,
		StopPrice: &stopPrice,
	}

	_, err := broker.SubmitOrder(req, time.Now())
	assert.NoError(t, err)

	candle := &market.Candle{
		Open:   50000.0,
		High:   50100.0,
		Low:    49900.0,
		Close:  50050.0,
		Volume: 100.0,
	}

	filled, err := broker.ProcessPendingOrders(candle, time.Now())
	assert.NoError(t, err)
	assert.Len(t, filled, 0)
}

func TestBrokerStopOrderTriggered(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	stopPrice := 49950.0 // Within candle range
	req := order.OrderRequest{
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      order.OrderSideSell,
		Type:      order.OrderTypeStop,
		Quantity:  1.0,
		StopPrice: &stopPrice,
	}

	ord, err := broker.SubmitOrder(req, time.Now())
	assert.NoError(t, err)

	candle := &market.Candle{
		Open:   50000.0,
		High:   50100.0,
		Low:    49900.0,
		Close:  50050.0,
		Volume: 100.0,
	}

	filled, err := broker.ProcessPendingOrders(candle, time.Now())
	assert.NoError(t, err)
	assert.Len(t, filled, 1)

	updated, _ := broker.GetOrder(ord.ID)
	assert.Equal(t, order.OrderStatusFilled, updated.Status)
}

func TestBrokerCancelOrder(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	price := 49000.0
	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeLimit,
		Quantity: 1.0,
		Price:    &price,
	}

	ord, err := broker.SubmitOrder(req, time.Now())
	assert.NoError(t, err)

	err = broker.CancelOrder(ord.ID, time.Now())
	assert.NoError(t, err)

	updated, _ := broker.GetOrder(ord.ID)
	assert.Equal(t, order.OrderStatusCancelled, updated.Status)

	pending := broker.GetPendingOrders()
	assert.Len(t, pending, 0)
}

func TestBrokerOrderHistory(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	req := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	_, err := broker.SubmitOrder(req, time.Now())
	assert.NoError(t, err)

	candle := &market.Candle{
		Open:   50000.0,
		High:   50100.0,
		Low:    49900.0,
		Close:  50050.0,
		Volume: 100.0,
	}

	_, err = broker.ProcessPendingOrders(candle, time.Now())
	assert.NoError(t, err)

	history := broker.GetOrderHistory()
	assert.GreaterOrEqual(t, len(history), 3) // created, accepted, filled
}

func TestBrokerGetOrdersBySymbol(t *testing.T) {
	config := execution.Config{
		SlippageType:  "percentage",
		SlippageValue: 0.0005,
		FeeMaker:      0.0002,
		FeeTaker:      0.0005,
		Seed:          42,
	}

	broker := NewBroker(execution.NewSimpleExecutor(config))

	req1 := order.OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	req2 := order.OrderRequest{
		Symbol:   market.Symbol("ETHUSDT"),
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 2.0,
	}

	_, err := broker.SubmitOrder(req1, time.Now())
	assert.NoError(t, err)

	_, err = broker.SubmitOrder(req2, time.Now())
	assert.NoError(t, err)

	btcOrders := broker.GetOrdersBySymbol(market.Symbol("BTCUSDT"))
	assert.Len(t, btcOrders, 1)

	ethOrders := broker.GetOrdersBySymbol(market.Symbol("ETHUSDT"))
	assert.Len(t, ethOrders, 1)
}
