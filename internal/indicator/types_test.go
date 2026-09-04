package indicator

import "testing"

func TestNewIndicatorRegistry(t *testing.T) {
	reg := NewIndicatorRegistry()

	if reg == nil {
		t.Fatal("expected registry, got nil")
	}

	if len(reg.List()) != 0 {
		t.Errorf("expected empty registry, got %d indicators", len(reg.List()))
	}
}

func TestIndicatorRegistry_Register(t *testing.T) {
	reg := NewIndicatorRegistry()

	// Mock indicator factory
	factory := func(params map[string]interface{}) Indicator {
		return nil
	}

	reg.Register("test_indicator", factory)

	if !reg.Exists("test_indicator") {
		t.Error("expected indicator to be registered")
	}

	list := reg.List()
	if len(list) != 1 {
		t.Errorf("expected 1 indicator, got %d", len(list))
	}
}
