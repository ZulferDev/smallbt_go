package broker

import (
	"context"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/execution"
	"github.com/ZulferDev/smallbt_go/internal/order"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
)

func TestPaperBroker_Interface(t *testing.T) {
	// Verify PaperBroker implements Broker interface
	var _ Broker = (*PaperBroker)(nil)
}

func TestPaperBroker_SubmitOrder(t *testing.T) {
	executor := execution.NewSimpleExecutor(execution.Config{})
	port := portfolio.NewPortfolio(10000)
	broker := NewPaperBroker(executor, port, DefaultLatencyConfig())
	defer broker.Close()

	ctx := context.Background()
	ord := &order.Order{
		Symbol:   "BTCUSDT",
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	// Submit order
	orderID, err := broker.SubmitOrder(ctx, ord)
	if err != nil {
		t.Fatalf("SubmitOrder failed: %v", err)
	}

	if orderID == "" {
		t.Error("Expected non-empty order ID")
	}

	// Check order is in queue
	if broker.orderQueue.Count() != 1 {
		t.Errorf("Expected 1 order in queue, got %d", broker.orderQueue.Count())
	}

	// Check order status is pending
	if broker.orderQueue.CountByStatus(StatusPending) != 1 {
		t.Errorf("Expected 1 pending order, got %d", broker.orderQueue.CountByStatus(StatusPending))
	}
}

func TestPaperBroker_OrderLifecycle(t *testing.T) {
	// Use fixed latency for predictable testing
	config := LatencyConfig{
		MinLatency: 100 * time.Millisecond,
		MaxLatency: 100 * time.Millisecond,
		Seed:       42,
	}

	executor := execution.NewSimpleExecutor(execution.Config{})
	port := portfolio.NewPortfolio(10000)
	broker := NewPaperBroker(executor, port, config)
	defer broker.Close()

	// Update price so order can fill
	broker.UpdatePrice("BTCUSDT", 50000.0)

	ctx := context.Background()
	ord := &order.Order{
		Symbol:   "BTCUSDT",
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 0.1,
	}

	// Submit order
	orderID, err := broker.SubmitOrder(ctx, ord)
	if err != nil {
		t.Fatalf("SubmitOrder failed: %v", err)
	}

	// Order should be pending
	qo, exists := broker.orderQueue.Get(orderID)
	if !exists {
		t.Fatal("Order not found in queue")
	}
	if qo.Status != StatusPending {
		t.Errorf("Expected pending status, got %s", qo.Status)
	}

	// Wait less than latency - should still be pending
	time.Sleep(50 * time.Millisecond)
	err = broker.ProcessOrderQueue(time.Now())
	if err != nil {
		t.Fatalf("ProcessOrderQueue failed: %v", err)
	}

	qo, _ = broker.orderQueue.Get(orderID)
	if qo.Status != StatusPending {
		t.Errorf("After 50ms, expected pending, got %s", qo.Status)
	}

	// Wait past latency - should be accepted or filled
	time.Sleep(60 * time.Millisecond)
	err = broker.ProcessOrderQueue(time.Now())
	if err != nil {
		t.Fatalf("ProcessOrderQueue failed: %v", err)
	}

	qo, _ = broker.orderQueue.Get(orderID)
	// Market orders may fill immediately after acceptance
	if qo.Status != StatusAccepted && qo.Status != StatusFilled {
		t.Errorf("After 110ms, expected accepted or filled, got %s", qo.Status)
	}

	// Process again if not yet filled
	if qo.Status == StatusAccepted {
		err = broker.ProcessOrderQueue(time.Now())
		if err != nil {
			t.Fatalf("ProcessOrderQueue failed: %v", err)
		}

		qo, _ = broker.orderQueue.Get(orderID)
		if qo.Status != StatusFilled {
			t.Errorf("After fill attempt, expected filled, got %s", qo.Status)
		}
	}
}

func TestPaperBroker_UpdatePrice(t *testing.T) {
	executor := execution.NewSimpleExecutor(execution.Config{})
	port := portfolio.NewPortfolio(10000)
	broker := NewPaperBroker(executor, port, DefaultLatencyConfig())
	defer broker.Close()

	// Update price
	broker.UpdatePrice("BTCUSDT", 50000.0)

	// Check price stored
	ctx := context.Background()
	price, err := broker.GetLastPrice(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("GetLastPrice failed: %v", err)
	}

	if price != 50000.0 {
		t.Errorf("Expected price 50000.0, got %.2f", price)
	}
}

func TestPaperBroker_CancelOrder(t *testing.T) {
	executor := execution.NewSimpleExecutor(execution.Config{})
	port := portfolio.NewPortfolio(10000)
	broker := NewPaperBroker(executor, port, DefaultLatencyConfig())
	defer broker.Close()

	ctx := context.Background()
	ord := &order.Order{
		Symbol:   "BTCUSDT",
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	// Submit order
	orderID, err := broker.SubmitOrder(ctx, ord)
	if err != nil {
		t.Fatalf("SubmitOrder failed: %v", err)
	}

	// Cancel order
	err = broker.CancelOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	// Order should be removed from queue
	_, exists := broker.orderQueue.Get(orderID)
	if exists {
		t.Error("Order still in queue after cancellation")
	}
}

func TestPaperBroker_Close(t *testing.T) {
	executor := execution.NewSimpleExecutor(execution.Config{})
	port := portfolio.NewPortfolio(10000)
	broker := NewPaperBroker(executor, port, DefaultLatencyConfig())

	// Close broker
	err := broker.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Operations after Close should fail
	ctx := context.Background()
	ord := &order.Order{
		ID:       "test",
		Symbol:   "BTCUSDT",
		Side:     order.OrderSideBuy,
		Type:     order.OrderTypeMarket,
		Quantity: 1.0,
	}

	_, err = broker.SubmitOrder(ctx, ord)
	if err != ErrBrokerClosed {
		t.Errorf("Expected ErrBrokerClosed, got %v", err)
	}
}

func TestLatencySimulator(t *testing.T) {
	config := LatencyConfig{
		MinLatency: 50 * time.Millisecond,
		MaxLatency: 200 * time.Millisecond,
		Seed:       42,
	}

	sim := NewLatencySimulator(config)

	// Test delay is within range
	for i := 0; i < 100; i++ {
		delay := sim.Delay()
		if delay < config.MinLatency || delay > config.MaxLatency {
			t.Errorf("Delay %v out of range [%v, %v]", delay, config.MinLatency, config.MaxLatency)
		}
	}
}

func TestLatencySimulator_FixedLatency(t *testing.T) {
	// When min == max, should return exact latency
	config := LatencyConfig{
		MinLatency: 100 * time.Millisecond,
		MaxLatency: 100 * time.Millisecond,
		Seed:       42,
	}

	sim := NewLatencySimulator(config)

	for i := 0; i < 10; i++ {
		delay := sim.Delay()
		if delay != 100*time.Millisecond {
			t.Errorf("Expected fixed delay 100ms, got %v", delay)
		}
	}
}

func TestLatencySimulator_OrderAcceptance(t *testing.T) {
	config := LatencyConfig{
		MinLatency: 100 * time.Millisecond,
		MaxLatency: 100 * time.Millisecond,
		Seed:       42,
	}

	sim := NewLatencySimulator(config)

	submitTime := time.Now()
	acceptTime := sim.SimulateOrderAcceptance(submitTime)

	expectedAcceptTime := submitTime.Add(100 * time.Millisecond)
	if !acceptTime.Equal(expectedAcceptTime) {
		t.Errorf("Expected accept time %v, got %v", expectedAcceptTime, acceptTime)
	}
}
