package execution

import (
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
)

func TestFixedSlippageModel(t *testing.T) {
	model := &FixedSlippageModel{
		Amount: 10.0,
	}

	candle := &market.Candle{
		Timestamp: time.Now(),
		Close:     50000,
	}

	t.Run("buy order", func(t *testing.T) {
		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Buy should have positive slippage (pay more)
		if slippage != 10.0 {
			t.Errorf("Buy slippage = %f, want 10.0", slippage)
		}
	})

	t.Run("sell order", func(t *testing.T) {
		req := order.OrderRequest{
			Side:     order.OrderSideSell,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Sell should have negative slippage (receive less)
		if slippage != -10.0 {
			t.Errorf("Sell slippage = %f, want -10.0", slippage)
		}
	})

	t.Run("name", func(t *testing.T) {
		if model.Name() != "fixed" {
			t.Errorf("Name() = %s, want fixed", model.Name())
		}
	})
}

func TestPercentageSlippageModel(t *testing.T) {
	model := &PercentageSlippageModel{
		Percentage: 0.001, // 0.1%
	}

	candle := &market.Candle{
		Timestamp: time.Now(),
		Close:     50000,
	}

	t.Run("buy order", func(t *testing.T) {
		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// 0.1% of 50000 = 50
		expected := 50.0
		if slippage != expected {
			t.Errorf("Buy slippage = %f, want %f", slippage, expected)
		}
	})

	t.Run("sell order", func(t *testing.T) {
		req := order.OrderRequest{
			Side:     order.OrderSideSell,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Negative for sell
		expected := -50.0
		if slippage != expected {
			t.Errorf("Sell slippage = %f, want %f", slippage, expected)
		}
	})
}

func TestVolatilitySlippageModel(t *testing.T) {
	model := &VolatilitySlippageModel{
		VolatilityFactor: 0.1, // 10% of range
		MinSlippage:      5.0,
		MaxSlippage:      500.0, // Increased to allow 200
	}

	t.Run("normal volatility", func(t *testing.T) {
		candle := &market.Candle{
			Timestamp: time.Now(),
			High:      51000,
			Low:       49000,
			Close:     50000,
		}

		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Range = 2000, 10% = 200
		expected := 200.0
		if slippage != expected {
			t.Errorf("Slippage = %f, want %f", slippage, expected)
		}
	})

	t.Run("low volatility clamped to min", func(t *testing.T) {
		candle := &market.Candle{
			Timestamp: time.Now(),
			High:      50010,
			Low:       49990,
			Close:     50000,
		}

		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Range = 20, 10% = 2, but min is 5
		if slippage != 5.0 {
			t.Errorf("Slippage = %f, want 5.0 (min)", slippage)
		}
	})

	t.Run("high volatility clamped to max", func(t *testing.T) {
		candle := &market.Candle{
			Timestamp: time.Now(),
			High:      55000,
			Low:       45000,
			Close:     50000,
		}

		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Range = 10000, 10% = 1000, but max is 500
		if slippage != 500.0 {
			t.Errorf("Slippage = %f, want 500.0 (max)", slippage)
		}
	})

	t.Run("zero range uses min", func(t *testing.T) {
		candle := &market.Candle{
			Timestamp: time.Now(),
			High:      50000,
			Low:       50000,
			Close:     50000,
		}

		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		if slippage != 5.0 {
			t.Errorf("Slippage = %f, want 5.0 (min)", slippage)
		}
	})

	t.Run("nil candle returns error", func(t *testing.T) {
		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0,
		}

		_, err := model.CalculateSlippage(req, 50000, nil)
		if err == nil {
			t.Error("Expected error for nil candle")
		}
	})
}

func TestVolumeSlippageModel(t *testing.T) {
	model := &VolumeSlippageModel{
		ImpactFactor: 0.01, // 1% impact per unit of volume ratio
		MinSlippage:  1.0,
		MaxSlippage:  500.0,
	}

	t.Run("small order low impact", func(t *testing.T) {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Close:     50000,
			Volume:    100.0, // 100 BTC volume
		}

		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0, // 1 BTC order
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Order is 1% of volume, impact = 0.01 * 0.01 * 50000 = 5
		expected := 5.0
		if slippage != expected {
			t.Errorf("Slippage = %f, want %f", slippage, expected)
		}
	})

	t.Run("large order high impact", func(t *testing.T) {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Close:     50000,
			Volume:    10.0, // 10 BTC volume
		}

		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 5.0, // 5 BTC order (50% of volume)
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Order is 50% of volume, impact = 0.5 * 0.01 * 50000 = 250
		expected := 250.0
		if slippage != expected {
			t.Errorf("Slippage = %f, want %f", slippage, expected)
		}
	})

	t.Run("huge order clamped to max", func(t *testing.T) {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Close:     50000,
			Volume:    1.0, // 1 BTC volume
		}

		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 10.0, // 10 BTC order (1000% of volume!)
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		// Would be huge, but clamped to max
		if slippage != 500.0 {
			t.Errorf("Slippage = %f, want 500.0 (max)", slippage)
		}
	})

	t.Run("zero volume uses min", func(t *testing.T) {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Close:     50000,
			Volume:    0,
		}

		req := order.OrderRequest{
			Side:     order.OrderSideBuy,
			Quantity: 1.0,
		}

		slippage, err := model.CalculateSlippage(req, 50000, candle)
		if err != nil {
			t.Fatalf("CalculateSlippage() error = %v", err)
		}

		if slippage != 1.0 {
			t.Errorf("Slippage = %f, want 1.0 (min)", slippage)
		}
	})
}

func TestNoSlippageModel(t *testing.T) {
	model := &NoSlippageModel{}

	candle := &market.Candle{
		Timestamp: time.Now(),
		Close:     50000,
	}

	req := order.OrderRequest{
		Side:     order.OrderSideBuy,
		Quantity: 1.0,
	}

	slippage, err := model.CalculateSlippage(req, 50000, candle)
	if err != nil {
		t.Fatalf("CalculateSlippage() error = %v", err)
	}

	if slippage != 0 {
		t.Errorf("Slippage = %f, want 0", slippage)
	}

	if model.Name() != "none" {
		t.Errorf("Name() = %s, want none", model.Name())
	}
}
