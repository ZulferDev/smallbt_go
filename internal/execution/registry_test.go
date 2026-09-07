package execution

import (
	"errors"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
)

// mockSlippageModel for testing.
type mockSlippageModel struct {
	name     string
	slippage float64
	err      error
}

func (m *mockSlippageModel) Name() string {
	return m.name
}

func (m *mockSlippageModel) CalculateSlippage(req order.OrderRequest, fillPrice float64, candle *market.Candle) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.slippage, nil
}

// TestNewRegistry tests registry creation.
func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}

	// Should have built-in models
	if !r.Has("fixed") {
		t.Fatal("Built-in 'fixed' model not registered")
	}
	if !r.Has("percentage") {
		t.Fatal("Built-in 'percentage' model not registered")
	}
	if !r.Has("volatility") {
		t.Fatal("Built-in 'volatility' model not registered")
	}
	if !r.Has("volume") {
		t.Fatal("Built-in 'volume' model not registered")
	}
	if !r.Has("none") {
		t.Fatal("Built-in 'none' model not registered")
	}
}

// TestRegister tests custom model registration.
func TestRegister(t *testing.T) {
	r := NewRegistry()

	factory := func(params map[string]interface{}) (SlippageModel, error) {
		return &mockSlippageModel{name: "custom", slippage: 0.5}, nil
	}

	err := r.Register("custom", factory)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !r.Has("custom") {
		t.Fatal("Custom model not found after registration")
	}
}

// TestRegisterNil tests registering nil factory.
func TestRegisterNil(t *testing.T) {
	r := NewRegistry()
	err := r.Register("test", nil)
	if err == nil {
		t.Fatal("Expected error when registering nil factory")
	}
}

// TestRegisterEmptyName tests registering with empty name.
func TestRegisterEmptyName(t *testing.T) {
	r := NewRegistry()
	factory := func(params map[string]interface{}) (SlippageModel, error) {
		return &mockSlippageModel{}, nil
	}

	err := r.Register("", factory)
	if err == nil {
		t.Fatal("Expected error when registering with empty name")
	}
}

// TestRegisterDuplicate tests registering duplicate model name.
func TestRegisterDuplicate(t *testing.T) {
	r := NewRegistry()

	factory := func(params map[string]interface{}) (SlippageModel, error) {
		return &mockSlippageModel{}, nil
	}

	r.Register("duplicate", factory)
	err := r.Register("duplicate", factory)

	if err == nil {
		t.Fatal("Expected error when registering duplicate model")
	}
}

// TestUnregister tests model unregistration.
func TestUnregister(t *testing.T) {
	r := NewRegistry()

	factory := func(params map[string]interface{}) (SlippageModel, error) {
		return &mockSlippageModel{}, nil
	}

	r.Register("custom", factory)
	if !r.Has("custom") {
		t.Fatal("Model not registered")
	}

	err := r.Unregister("custom")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	if r.Has("custom") {
		t.Fatal("Model still exists after unregister")
	}
}

// TestUnregisterBuiltin tests that built-in models cannot be unregistered.
func TestUnregisterBuiltin(t *testing.T) {
	r := NewRegistry()

	err := r.Unregister("fixed")
	if err == nil {
		t.Fatal("Expected error when unregistering built-in model")
	}

	// Should still exist
	if !r.Has("fixed") {
		t.Fatal("Built-in model was removed")
	}
}

// TestGet tests retrieving factory.
func TestGet(t *testing.T) {
	r := NewRegistry()

	factory := r.Get("fixed")
	if factory == nil {
		t.Fatal("Get returned nil for built-in model")
	}
}

// TestGetNonExistent tests retrieving non-existent factory.
func TestGetNonExistent(t *testing.T) {
	r := NewRegistry()
	factory := r.Get("nonexistent")
	if factory != nil {
		t.Fatal("Expected nil for non-existent model")
	}
}

// TestHas tests checking model existence.
func TestHas(t *testing.T) {
	r := NewRegistry()

	if !r.Has("fixed") {
		t.Fatal("Has returned false for existing model")
	}

	if r.Has("nonexistent") {
		t.Fatal("Has returned true for non-existent model")
	}
}

// TestList tests listing all models.
func TestList(t *testing.T) {
	r := NewRegistry()

	names := r.List()
	if len(names) < 5 {
		t.Fatalf("Expected at least 5 built-in models, got %d", len(names))
	}

	// Check all built-ins are present
	builtins := map[string]bool{
		"fixed": false, "percentage": false, "volatility": false,
		"volume": false, "none": false,
	}

	for _, name := range names {
		if _, ok := builtins[name]; ok {
			builtins[name] = true
		}
	}

	for name, found := range builtins {
		if !found {
			t.Fatalf("Built-in model %q not in list", name)
		}
	}
}

// TestCreate tests creating model instances.
func TestCreate(t *testing.T) {
	r := NewRegistry()

	// Test fixed model
	model, err := r.Create("fixed", map[string]interface{}{
		"amount": 1.0,
	})
	if err != nil {
		t.Fatalf("Create fixed model failed: %v", err)
	}
	if model == nil {
		t.Fatal("Created model is nil")
	}
	if model.Name() != "fixed" {
		t.Fatalf("Expected name 'fixed', got %q", model.Name())
	}
}

// TestCreatePercentage tests creating percentage model.
func TestCreatePercentage(t *testing.T) {
	r := NewRegistry()

	model, err := r.Create("percentage", map[string]interface{}{
		"percentage": 0.001,
	})
	if err != nil {
		t.Fatalf("Create percentage model failed: %v", err)
	}
	if model.Name() != "percentage" {
		t.Fatalf("Expected name 'percentage', got %q", model.Name())
	}
}

// TestCreateVolatility tests creating volatility model.
func TestCreateVolatility(t *testing.T) {
	r := NewRegistry()

	model, err := r.Create("volatility", map[string]interface{}{
		"factor":       0.1,
		"max_slippage": 10.0,
	})
	if err != nil {
		t.Fatalf("Create volatility model failed: %v", err)
	}
	if model.Name() != "volatility" {
		t.Fatalf("Expected name 'volatility', got %q", model.Name())
	}
}

// TestCreateVolume tests creating volume model.
func TestCreateVolume(t *testing.T) {
	r := NewRegistry()

	model, err := r.Create("volume", map[string]interface{}{
		"impact_factor": 1.0,
		"max_slippage":  5.0,
	})
	if err != nil {
		t.Fatalf("Create volume model failed: %v", err)
	}
	if model.Name() != "volume" {
		t.Fatalf("Expected name 'volume', got %q", model.Name())
	}
}

// TestCreateNone tests creating no-slippage model.
func TestCreateNone(t *testing.T) {
	r := NewRegistry()

	model, err := r.Create("none", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Create none model failed: %v", err)
	}
	if model.Name() != "none" {
		t.Fatalf("Expected name 'none', got %q", model.Name())
	}
}

// TestCreateNonExistent tests creating non-existent model.
func TestCreateNonExistent(t *testing.T) {
	r := NewRegistry()

	_, err := r.Create("nonexistent", nil)
	if err == nil {
		t.Fatal("Expected error when creating non-existent model")
	}
}

// TestCreateInvalidParams tests creating model with invalid parameters.
func TestCreateInvalidParams(t *testing.T) {
	r := NewRegistry()

	// Fixed model requires 'amount'
	_, err := r.Create("fixed", map[string]interface{}{
		"wrong_param": 1.0,
	})
	if err == nil {
		t.Fatal("Expected error with invalid parameters")
	}
}

// TestClear tests clearing custom models.
func TestClear(t *testing.T) {
	r := NewRegistry()

	// Add custom model
	factory := func(params map[string]interface{}) (SlippageModel, error) {
		return &mockSlippageModel{}, nil
	}
	r.Register("custom", factory)

	if !r.Has("custom") {
		t.Fatal("Custom model not registered")
	}

	// Clear
	r.Clear()

	// Custom should be gone
	if r.Has("custom") {
		t.Fatal("Custom model still exists after clear")
	}

	// Built-ins should remain
	if !r.Has("fixed") {
		t.Fatal("Built-in model removed after clear")
	}
}

// TestGlobalRegistry tests package-level functions.
func TestGlobalRegistry(t *testing.T) {
	// Has
	if !Has("fixed") {
		t.Fatal("Global Has returned false for built-in")
	}

	// Get
	factory := Get("fixed")
	if factory == nil {
		t.Fatal("Global Get returned nil")
	}

	// List
	names := List()
	if len(names) < 5 {
		t.Fatalf("Global List returned too few models: %d", len(names))
	}

	// Create
	model, err := Create("fixed", map[string]interface{}{"amount": 1.0})
	if err != nil {
		t.Fatalf("Global Create failed: %v", err)
	}
	if model == nil {
		t.Fatal("Global Create returned nil model")
	}
}

// TestGlobalRegisterUnregister tests global register/unregister.
func TestGlobalRegisterUnregister(t *testing.T) {
	factory := func(params map[string]interface{}) (SlippageModel, error) {
		return &mockSlippageModel{name: "global_test"}, nil
	}

	// Register
	err := Register("global_test", factory)
	if err != nil {
		t.Fatalf("Global Register failed: %v", err)
	}

	if !Has("global_test") {
		t.Fatal("Model not found after global Register")
	}

	// Unregister
	err = Unregister("global_test")
	if err != nil {
		t.Fatalf("Global Unregister failed: %v", err)
	}

	if Has("global_test") {
		t.Fatal("Model still exists after global Unregister")
	}
}

// TestGlobalRegistryInstance tests GlobalRegistry() function.
func TestGlobalRegistryInstance(t *testing.T) {
	registry := GlobalRegistry()
	if registry == nil {
		t.Fatal("GlobalRegistry() returned nil")
	}
	if registry != globalRegistry {
		t.Fatal("GlobalRegistry() returned different instance")
	}
}

// TestCustomModelCreation tests creating and using a custom model.
func TestCustomModelCreation(t *testing.T) {
	r := NewRegistry()

	// Register custom model
	err := r.Register("test_custom", func(params map[string]interface{}) (SlippageModel, error) {
		slippage, ok := params["slippage"].(float64)
		if !ok {
			return nil, errors.New("slippage parameter required")
		}
		return &mockSlippageModel{name: "test_custom", slippage: slippage}, nil
	})

	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Create instance
	model, err := r.Create("test_custom", map[string]interface{}{
		"slippage": 2.5,
	})

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test the model
	req := order.OrderRequest{
		Side:     order.OrderSideBuy,
		Quantity: 1.0,
	}

	slippage, err := model.CalculateSlippage(req, 100.0, nil)
	if err != nil {
		t.Fatalf("CalculateSlippage failed: %v", err)
	}

	if slippage != 2.5 {
		t.Fatalf("Expected slippage 2.5, got %f", slippage)
	}
}

// TestSpreadSlippageModel tests the example spread model.
func TestSpreadSlippageModel(t *testing.T) {
	model := NewSpreadSlippageModel(0.5)

	if model.Name() != "spread" {
		t.Fatalf("Expected name 'spread', got %q", model.Name())
	}

	candle := &market.Candle{
		High: 105.0,
		Low:  95.0,
	}

	req := order.OrderRequest{
		Side:     order.OrderSideBuy,
		Quantity: 1.0,
	}

	slippage, err := model.CalculateSlippage(req, 100.0, candle)
	if err != nil {
		t.Fatalf("CalculateSlippage failed: %v", err)
	}

	// Expected: (105-95) * 0.5 / 2 = 2.5
	expected := 2.5
	if slippage != expected {
		t.Fatalf("Expected slippage %.2f, got %.2f", expected, slippage)
	}
}

// TestRegisterSpreadModel tests registering the spread model.
func TestRegisterSpreadModel(t *testing.T) {
	r := NewRegistry()

	// Register spread model
	err := r.Register("spread", func(params map[string]interface{}) (SlippageModel, error) {
		multiplier, ok := params["multiplier"].(float64)
		if !ok {
			multiplier = 0.5
		}
		return NewSpreadSlippageModel(multiplier), nil
	})

	if err != nil {
		t.Fatalf("Register spread model failed: %v", err)
	}

	// Create instance
	model, err := r.Create("spread", map[string]interface{}{
		"multiplier": 0.3,
	})

	if err != nil {
		t.Fatalf("Create spread model failed: %v", err)
	}

	if model.Name() != "spread" {
		t.Fatalf("Expected name 'spread', got %q", model.Name())
	}
}
