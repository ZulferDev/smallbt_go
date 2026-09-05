package feed

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// mockCandleServer creates a WebSocket server that sends candle data.
func mockCandleServer(candles []map[string]interface{}) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send candles
		for _, candle := range candles {
			data, _ := json.Marshal(candle)
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}

		// Keep connection alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	return httptest.NewServer(handler)
}

func TestWebSocketFeed_ParseMessage_Valid(t *testing.T) {
	config := DefaultWebSocketConfig()
	feed := NewWebSocketFeed(config)

	message := []byte(`{
		"timestamp": 1609459200,
		"open": 29000.0,
		"high": 29500.0,
		"low": 28500.0,
		"close": 29200.0,
		"volume": 1000.0
	}`)

	candle, err := feed.parseMessage(message)
	if err != nil {
		t.Fatalf("parseMessage() failed: %v", err)
	}

	if candle == nil {
		t.Fatal("expected candle to be parsed")
	}

	if candle.Open != 29000.0 {
		t.Errorf("expected open 29000.0, got %f", candle.Open)
	}

	if candle.High != 29500.0 {
		t.Errorf("expected high 29500.0, got %f", candle.High)
	}

	if candle.Low != 28500.0 {
		t.Errorf("expected low 28500.0, got %f", candle.Low)
	}

	if candle.Close != 29200.0 {
		t.Errorf("expected close 29200.0, got %f", candle.Close)
	}

	if candle.Volume != 1000.0 {
		t.Errorf("expected volume 1000.0, got %f", candle.Volume)
	}
}

func TestWebSocketFeed_ParseMessage_Invalid(t *testing.T) {
	config := DefaultWebSocketConfig()
	feed := NewWebSocketFeed(config)

	tests := []struct {
		name    string
		message string
	}{
		{"invalid json", `{invalid json}`},
		{"missing fields", `{"timestamp": 1609459200, "open": 100, "high": 50, "low": 95, "close": 100, "volume": 100}`},
		{"invalid candle", `{"timestamp": 1609459200, "open": 100, "high": 90, "low": 95, "close": 100, "volume": 100}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := feed.parseMessage([]byte(tt.message))
			if err == nil {
				t.Error("expected error for invalid message")
			}
		})
	}
}

func TestWebSocketFeed_ReceiveCandles(t *testing.T) {
	candles := []map[string]interface{}{
		{
			"timestamp": int64(1609459200),
			"open":      29000.0,
			"high":      29500.0,
			"low":       28500.0,
			"close":     29200.0,
			"volume":    1000.0,
		},
		{
			"timestamp": int64(1609459260),
			"open":      29200.0,
			"high":      29600.0,
			"low":       28800.0,
			"close":     29400.0,
			"volume":    1200.0,
		},
	}

	server := mockCandleServer(candles)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	config := DefaultWebSocketConfig()
	config.URL = wsURL

	feed := NewWebSocketFeed(config)

	err := feed.Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer feed.Close()

	ch := feed.Subscribe()

	// Receive candles
	received := make([]*market.Candle, 0)
	timeout := time.After(500 * time.Millisecond)

	for i := 0; i < 2; i++ {
		select {
		case candle := <-ch:
			if candle != nil {
				received = append(received, candle)
			}
		case <-timeout:
			t.Fatalf("timeout waiting for candle %d", i)
		}
	}

	if len(received) != 2 {
		t.Errorf("expected 2 candles, got %d", len(received))
	}

	// Verify first candle
	if received[0].Open != 29000.0 {
		t.Errorf("candle 0: expected open 29000.0, got %f", received[0].Open)
	}

	// Verify second candle
	if received[1].Open != 29200.0 {
		t.Errorf("candle 1: expected open 29200.0, got %f", received[1].Open)
	}
}

func TestWebSocketFeed_BufferDrain(t *testing.T) {
	config := DefaultWebSocketConfig()
	config.BufferSize = 3
	feed := NewWebSocketFeed(config)

	// Push candles to buffer
	for i := 0; i < 5; i++ {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Open:      float64(100 + i),
			High:      float64(105 + i),
			Low:       float64(95 + i),
			Close:     float64(102 + i),
			Volume:    1000.0,
		}
		feed.buffer.Push(candle)
	}

	// Buffer should trigger overflow at 3
	if feed.buffer.Len() > 3 {
		t.Errorf("expected buffer len <= 3, got %d", feed.buffer.Len())
	}
}

func TestWebSocketFeed_MultipleSubscribers(t *testing.T) {
	candles := []map[string]interface{}{
		{
			"timestamp": int64(1609459200),
			"open":      29000.0,
			"high":      29500.0,
			"low":       28500.0,
			"close":     29200.0,
			"volume":    1000.0,
		},
	}

	server := mockCandleServer(candles)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	config := DefaultWebSocketConfig()
	config.URL = wsURL

	feed := NewWebSocketFeed(config)

	err := feed.Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer feed.Close()

	// Create multiple subscribers
	ch1 := feed.Subscribe()
	ch2 := feed.Subscribe()
	ch3 := feed.Subscribe()

	timeout := time.After(500 * time.Millisecond)

	// All subscribers should receive the candle
	for i, ch := range []<-chan *market.Candle{ch1, ch2, ch3} {
		select {
		case candle := <-ch:
			if candle == nil {
				t.Errorf("subscriber %d: received nil candle", i)
			}
		case <-timeout:
			t.Errorf("subscriber %d: timeout waiting for candle", i)
		}
	}
}

func TestWebSocketFeed_BufferPersistence(t *testing.T) {
	config := DefaultWebSocketConfig()
	config.BufferSize = 100
	feed := NewWebSocketFeed(config)

	// Add candles to buffer
	for i := 0; i < 10; i++ {
		candle := &market.Candle{
			Timestamp: time.Now(),
			Open:      float64(100 + i),
			High:      float64(105 + i),
			Low:       float64(95 + i),
			Close:     float64(102 + i),
			Volume:    1000.0,
		}
		feed.buffer.Push(candle)
	}

	if feed.buffer.Len() != 10 {
		t.Errorf("expected buffer len 10, got %d", feed.buffer.Len())
	}

	// Drain should return all candles
	drained := feed.buffer.Drain()
	if len(drained) != 10 {
		t.Errorf("expected 10 drained candles, got %d", len(drained))
	}

	if feed.buffer.Len() != 0 {
		t.Errorf("expected buffer len 0 after drain, got %d", feed.buffer.Len())
	}
}
