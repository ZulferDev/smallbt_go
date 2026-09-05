package broker

import (
	"context"
	"errors"

	"github.com/ZulferDev/smallbt_go/internal/order"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

// Broker abstracts order execution (simulated, paper, or live)
// This interface allows the same strategy code to run in different execution modes
type Broker interface {
	// SubmitOrder submits an order for execution
	// Returns order ID on success
	SubmitOrder(ctx context.Context, o *order.Order) (string, error)

	// CancelOrder attempts to cancel an order
	// Returns error if order cannot be cancelled (already filled, etc)
	CancelOrder(ctx context.Context, orderID string) error

	// GetOrder retrieves order status by ID
	GetOrder(ctx context.Context, orderID string) (*order.Order, error)

	// GetPositions returns all open positions
	GetPositions(ctx context.Context) ([]*portfolio.Position, error)

	// GetBalance returns current account balance
	GetBalance(ctx context.Context) (*portfolio.Balance, error)

	// GetLastPrice returns the most recent market price for a symbol
	// Used for order validation and risk checks
	GetLastPrice(ctx context.Context, symbol string) (float64, error)

	// Close releases resources (WebSocket connections, etc)
	Close() error
}

var (
	// ErrOrderNotFound is returned when order ID doesn't exist
	ErrOrderNotFound = errors.New("order not found")

	// ErrInsufficientBalance is returned when account has insufficient funds
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrInvalidPrice is returned when order price is invalid
	ErrInvalidPrice = errors.New("invalid price")

	// ErrInvalidQuantity is returned when order quantity is invalid
	ErrInvalidQuantity = errors.New("invalid quantity")

	// ErrOrderAlreadyFilled is returned when trying to cancel a filled order
	ErrOrderAlreadyFilled = errors.New("order already filled")

	// ErrOrderAlreadyCancelled is returned when trying to cancel a cancelled order
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")

	// ErrSymbolNotFound is returned when symbol doesn't exist
	ErrSymbolNotFound = errors.New("symbol not found")

	// ErrBrokerClosed is returned when operation is attempted on closed broker
	ErrBrokerClosed = errors.New("broker is closed")
)
