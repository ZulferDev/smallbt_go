package feed

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// mockWebSocketServer creates a test WebSocket server.
func mockWebSocketServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo loop for testing
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				return
			}
		}
	})
	
	return httptest.NewServer(handler)
}

func TestWebSocketFeed_NewWebSocketFeed(t *testing.T) {
	config := DefaultWebSocketConfig()
	config.URL = "ws://localhost:8080"
	config.Symbols = []string{"BTCUSDT"}
	config.Timeframe = time.Hour
	
	feed := NewWebSocketFeed(config)
	
	if feed == nil {
		t.Fatal("expected feed to be created")
	}
	
	if feed.State() != StateDisconnected {
		t.Errorf("expected initial state to be disconnected, got %s", feed.State())
	}
	
	if feed.url != config.URL {
		t.Errorf("expected url %s, got %s", config.URL, feed.url)
	}
	
	if len(feed.symbols) != 1 || feed.symbols[0] != "BTCUSDT" {
		t.Errorf("expected symbols [BTCUSDT], got %v", feed.symbols)
	}
}

func TestWebSocketFeed_Connect(t *testing.T) {
	server := mockWebSocketServer()
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	config := DefaultWebSocketConfig()
	config.URL = wsURL
	config.Symbols = []string{"BTCUSDT"}
	config.Timeframe = time.Hour
	
	feed := NewWebSocketFeed(config)
	
	err := feed.Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	
	// Wait for state to update
	time.Sleep(50 * time.Millisecond)
	
	if feed.State() != StateConnected {
		t.Errorf("expected state connected, got %s", feed.State())
	}
	
	// Cleanup
	err = feed.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestWebSocketFeed_ConnectInvalidURL(t *testing.T) {
	config := DefaultWebSocketConfig()
	config.URL = "ws://invalid-url-that-does-not-exist:9999"
	config.Symbols = []string{"BTCUSDT"}
	
	feed := NewWebSocketFeed(config)
	
	err := feed.Connect()
	if err == nil {
		t.Error("expected Connect() to fail with invalid URL")
		feed.Close()
	}
	
	if feed.State() == StateConnected {
		t.Error("expected state not to be connected")
	}
}

func TestWebSocketFeed_ConnectTwice(t *testing.T) {
	server := mockWebSocketServer()
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	config := DefaultWebSocketConfig()
	config.URL = wsURL
	
	feed := NewWebSocketFeed(config)
	
	err := feed.Connect()
	if err != nil {
		t.Fatalf("first Connect() failed: %v", err)
	}
	defer feed.Close()
	
	time.Sleep(50 * time.Millisecond)
	
	// Try connecting again
	err = feed.Connect()
	if err == nil {
		t.Error("expected second Connect() to fail")
	}
}

func TestWebSocketFeed_Close(t *testing.T) {
	server := mockWebSocketServer()
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	config := DefaultWebSocketConfig()
	config.URL = wsURL
	
	feed := NewWebSocketFeed(config)
	
	err := feed.Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	
	time.Sleep(50 * time.Millisecond)
	
	err = feed.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
	
	if feed.State() != StateClosed {
		t.Errorf("expected state closed, got %s", feed.State())
	}
	
	// Close again should not error
	err = feed.Close()
	if err != nil {
		t.Errorf("second Close() failed: %v", err)
	}
}

func TestWebSocketFeed_Subscribe(t *testing.T) {
	config := DefaultWebSocketConfig()
	config.URL = "ws://localhost:8080"
	
	feed := NewWebSocketFeed(config)
	
	ch := feed.Subscribe()
	if ch == nil {
		t.Fatal("expected subscription channel to be created")
	}
	
	// Subscribe again
	ch2 := feed.Subscribe()
	if ch2 == nil {
		t.Fatal("expected second subscription channel to be created")
	}
	
	if ch == ch2 {
		t.Error("expected different channels for different subscriptions")
	}
	
	feed.Close()
}

func TestConnectionState_String(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateDisconnected, "disconnected"},
		{StateConnecting, "connecting"},
		{StateConnected, "connected"},
		{StateReconnecting, "reconnecting"},
		{StateClosed, "closed"},
		{ConnectionState(999), "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDefaultWebSocketConfig(t *testing.T) {
	config := DefaultWebSocketConfig()
	
	if config.ReconnectDelay != 1*time.Second {
		t.Errorf("expected ReconnectDelay 1s, got %v", config.ReconnectDelay)
	}
	
	if config.MaxReconnects != 10 {
		t.Errorf("expected MaxReconnects 10, got %d", config.MaxReconnects)
	}
	
	if config.PingInterval != 30*time.Second {
		t.Errorf("expected PingInterval 30s, got %v", config.PingInterval)
	}
	
	if config.PongTimeout != 10*time.Second {
		t.Errorf("expected PongTimeout 10s, got %v", config.PongTimeout)
	}
	
	if config.BufferSize != 1000 {
		t.Errorf("expected BufferSize 1000, got %d", config.BufferSize)
	}
}
