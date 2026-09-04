package order

import (
	"testing"
	"time"

	"github.com/1jehuang/backtest/internal/market"
	"github.com/stretchr/testify/assert"
)

func TestOrderManagerCreateOrder(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, err := om.CreateOrder(req, time.Now())
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, OrderStatusPending, order.Status)
	assert.Equal(t, 1.0, order.Quantity)
	assert.Equal(t, OrderSideBuy, order.Side)
	assert.Equal(t, 1, om.GetOrderCount())
	assert.Equal(t, 1, om.GetEventCount())
}

func TestOrderManagerCreateLimitOrderWithoutPrice(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeLimit,
		Quantity: 1.0,
	}

	order, err := om.CreateOrder(req, time.Now())
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "requires price")
}

func TestOrderManagerCreateStopOrderWithoutStopPrice(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeStop,
		Quantity: 1.0,
	}

	order, err := om.CreateOrder(req, time.Now())
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "requires stop price")
}

func TestOrderManagerCreateStopLimitOrderWithoutPrices(t *testing.T) {
	om := NewOrderManager()

	// Missing both prices
	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeStopLimit,
		Quantity: 1.0,
	}

	order, err := om.CreateOrder(req, time.Now())
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "requires both price and stop price")

	// Missing stop price
	price := 50000.0
	req = OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeStopLimit,
		Quantity: 1.0,
		Price:    &price,
	}

	order, err = om.CreateOrder(req, time.Now())
	assert.Error(t, err)
	assert.Nil(t, order)
}

func TestOrderManagerAcceptOrder(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())
	assert.Equal(t, OrderStatusPending, order.Status)

	err := om.AcceptOrder(order.ID, time.Now())
	assert.NoError(t, err)

	updated, _ := om.GetOrder(order.ID)
	assert.Equal(t, OrderStatusAccepted, updated.Status)
	assert.Equal(t, 2, om.GetEventCount())
}

func TestOrderManagerFillOrder(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())
	_ = om.AcceptOrder(order.ID, time.Now())

	fill := &Fill{
		OrderID:   order.ID,
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      OrderSideBuy,
		Quantity:  1.0,
		Price:     50000.0,
		Timestamp: time.Now(),
		Fees:      25.0,
		Slippage:  10.0,
	}

	err := om.FillOrder(order.ID, fill)
	assert.NoError(t, err)

	updated, _ := om.GetOrder(order.ID)
	assert.Equal(t, OrderStatusFilled, updated.Status)
	assert.Equal(t, 1.0, updated.FilledQty)
	assert.Equal(t, 50000.0, updated.FilledPrice)
	assert.Equal(t, 25.0, updated.Fees)
}

func TestOrderManagerPartialFill(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 2.0,
	}

	order, _ := om.CreateOrder(req, time.Now())
	_ = om.AcceptOrder(order.ID, time.Now())

	// First partial fill
	fill1 := &Fill{
		OrderID:   order.ID,
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      OrderSideBuy,
		Quantity:  1.0,
		Price:     50000.0,
		Timestamp: time.Now(),
		Fees:      25.0,
	}

	err := om.FillOrder(order.ID, fill1)
	assert.NoError(t, err)

	updated, _ := om.GetOrder(order.ID)
	assert.Equal(t, OrderStatusPartiallyFilled, updated.Status)
	assert.Equal(t, 1.0, updated.FilledQty)

	// Second fill completes the order
	fill2 := &Fill{
		OrderID:   order.ID,
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      OrderSideBuy,
		Quantity:  1.0,
		Price:     50100.0,
		Timestamp: time.Now(),
		Fees:      25.0,
	}

	err = om.FillOrder(order.ID, fill2)
	assert.NoError(t, err)

	updated, _ = om.GetOrder(order.ID)
	assert.Equal(t, OrderStatusFilled, updated.Status)
	assert.Equal(t, 2.0, updated.FilledQty)
	// Weighted average price: (50000*1 + 50100*1) / 2 = 50050
	assert.Equal(t, 50050.0, updated.FilledPrice)
}

func TestOrderManagerRejectOrder(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())

	err := om.RejectOrder(order.ID, "insufficient funds", time.Now())
	assert.NoError(t, err)

	updated, _ := om.GetOrder(order.ID)
	assert.Equal(t, OrderStatusRejected, updated.Status)
}

func TestOrderManagerCancelOrder(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())
	_ = om.AcceptOrder(order.ID, time.Now())

	err := om.CancelOrder(order.ID, time.Now())
	assert.NoError(t, err)

	updated, _ := om.GetOrder(order.ID)
	assert.Equal(t, OrderStatusCancelled, updated.Status)
}

func TestOrderManagerExpireOrder(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())

	err := om.ExpireOrder(order.ID, time.Now())
	assert.NoError(t, err)

	updated, _ := om.GetOrder(order.ID)
	assert.Equal(t, OrderStatusExpired, updated.Status)
}

func TestOrderManagerGetOrdersBySymbol(t *testing.T) {
	om := NewOrderManager()

	req1 := OrderRequest{Symbol: market.Symbol("BTCUSDT"), Side: OrderSideBuy, Type: OrderTypeMarket, Quantity: 1.0}
	req2 := OrderRequest{Symbol: market.Symbol("ETHUSDT"), Side: OrderSideBuy, Type: OrderTypeMarket, Quantity: 2.0}
	req3 := OrderRequest{Symbol: market.Symbol("BTCUSDT"), Side: OrderSideSell, Type: OrderTypeMarket, Quantity: 0.5}

	_, _ = om.CreateOrder(req1, time.Now())
	_, _ = om.CreateOrder(req2, time.Now())
	_, _ = om.CreateOrder(req3, time.Now())

	btcOrders := om.GetOrdersBySymbol(market.Symbol("BTCUSDT"))
	assert.Len(t, btcOrders, 2)

	ethOrders := om.GetOrdersBySymbol(market.Symbol("ETHUSDT"))
	assert.Len(t, ethOrders, 1)
}

func TestOrderManagerGetOrdersByStatus(t *testing.T) {
	om := NewOrderManager()

	req1 := OrderRequest{Symbol: market.Symbol("BTCUSDT"), Side: OrderSideBuy, Type: OrderTypeMarket, Quantity: 1.0}
	req2 := OrderRequest{Symbol: market.Symbol("ETHUSDT"), Side: OrderSideBuy, Type: OrderTypeMarket, Quantity: 2.0}

	order1, _ := om.CreateOrder(req1, time.Now())
	_, _ = om.CreateOrder(req2, time.Now())

	_ = om.AcceptOrder(order1.ID, time.Now())

	pendingOrders := om.GetOrdersByStatus(OrderStatusPending)
	assert.Len(t, pendingOrders, 1)

	acceptedOrders := om.GetOrdersByStatus(OrderStatusAccepted)
	assert.Len(t, acceptedOrders, 1)
}

func TestOrderManagerHistory(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())
	_ = om.AcceptOrder(order.ID, time.Now())

	fill := &Fill{
		OrderID:   order.ID,
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      OrderSideBuy,
		Quantity:  1.0,
		Price:     50000.0,
		Timestamp: time.Now(),
		Fees:      25.0,
	}
	_ = om.FillOrder(order.ID, fill)

	history := om.GetHistory()
	assert.Len(t, history, 3) // created, accepted, filled
	assert.Equal(t, "created", history[0].Type)
	assert.Equal(t, "accepted", history[1].Type)
	assert.Equal(t, "filled", history[2].Type)
}

func TestOrderManagerCannotFillCancelled(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())
	_ = om.CancelOrder(order.ID, time.Now())

	fill := &Fill{
		OrderID:   order.ID,
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      OrderSideBuy,
		Quantity:  1.0,
		Price:     50000.0,
		Timestamp: time.Now(),
	}

	err := om.FillOrder(order.ID, fill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

func TestOrderManagerCannotCancelFilled(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())
	_ = om.AcceptOrder(order.ID, time.Now())

	fill := &Fill{
		OrderID:   order.ID,
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      OrderSideBuy,
		Quantity:  1.0,
		Price:     50000.0,
		Timestamp: time.Now(),
	}
	_ = om.FillOrder(order.ID, fill)

	err := om.CancelOrder(order.ID, time.Now())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel")
}

func TestOrderManagerInvalidQuantity(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: -1.0,
	}

	order, err := om.CreateOrder(req, time.Now())
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "invalid quantity")
}

func TestOrderManagerFillExceedsQuantity(t *testing.T) {
	om := NewOrderManager()

	req := OrderRequest{
		Symbol:   market.Symbol("BTCUSDT"),
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 1.0,
	}

	order, _ := om.CreateOrder(req, time.Now())
	_ = om.AcceptOrder(order.ID, time.Now())

	fill := &Fill{
		OrderID:   order.ID,
		Symbol:    market.Symbol("BTCUSDT"),
		Side:      OrderSideBuy,
		Quantity:  2.0, // Exceeds order quantity
		Price:     50000.0,
		Timestamp: time.Now(),
	}

	err := om.FillOrder(order.ID, fill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds order quantity")
}
