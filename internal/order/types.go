package order

import (
	"time"

	"github.com/1jehuang/backtest/internal/market"
)

// OrderType represents the type of an order.
type OrderType string

const (
	OrderTypeMarket    OrderType = "market"
	OrderTypeLimit     OrderType = "limit"
	OrderTypeStop      OrderType = "stop"
	OrderTypeStopLimit OrderType = "stop_limit"
)

// OrderSide represents the side of an order.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusAccepted        OrderStatus = "accepted"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCancelled       OrderStatus = "cancelled"
	OrderStatusRejected        OrderStatus = "rejected"
	OrderStatusExpired         OrderStatus = "expired"
)

// Order represents a trading order.
type Order struct {
	ID          string
	Symbol      market.Symbol
	Side        OrderSide
	Type        OrderType
	Quantity    float64
	Price       *float64 // For limit orders
	StopPrice   *float64 // For stop orders
	Status      OrderStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	FilledQty   float64
	FilledPrice float64
	Fees        float64
}

// OrderRequest represents a request to create an order.
type OrderRequest struct {
	Symbol    market.Symbol
	Side      OrderSide
	Type      OrderType
	Quantity  float64
	Price     *float64
	StopPrice *float64
}

// Fill represents an order fill event.
type Fill struct {
	OrderID   string
	Symbol    market.Symbol
	Side      OrderSide
	Quantity  float64
	Price     float64
	Timestamp time.Time
	Fees      float64
	Slippage  float64
}
