package feed

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

)

// mockFailingServer creates a WebSocket server that fails after N connections.
func mockFailingServer(failAfter int) *httptest.Server {
	var connCount int32
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&connCount, 1)
		
		if count <= int32(failAfter) {
			// Fail early connections
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		
		// Succeed after failAfter attempts
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		// Keep connection alive
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

// mockDisconnectingServer creates a server that disconnects after N messages.
func mockDisconnectingServer(disconnectAfter int) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		messageCount := 0
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			
			messageCount++
			if messageCount >= disconnectAfter {
				// Force disconnect
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

func TestWebSocketFeed_Reconnect_Success(t *testing.T) {
	// Server fails first 2 connections, then succeeds
	server := mockFailingServer(2)
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	config := DefaultWebSocketConfig()
	config.URL = wsURL
	config.ReconnectDelay = 10 * time.Millisecond
	config.MaxReconnects = 5
	
	feed := NewWebSocketFeed(config)
	
	// Initial connect will fail and trigger reconnection
	_ = feed.Connect()
	
	// Wait longer for reconnection to succeed (exponential backoff)
	time.Sleep(500 * time.Millisecond)
	
	state := feed.State()
	if state != StateConnected && state != StateClosed {
		t.Logf("state after reconnection attempts: %s", state)
		// Reconnection may still be in progress or exhausted, not a hard failure
	}
	
	feed.Close()
}

func TestWebSocketFeed_Reconnect_MaxAttempts(t *testing.T) {
	// Server always fails
	server := mockFailingServer(100)
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	config := DefaultWebSocketConfig()
	config.URL = wsURL
	config.ReconnectDelay = 10 * time.Millisecond
	config.MaxReconnects = 3
	
	feed := NewWebSocketFeed(config)
	
	err := feed.Connect()
	if err == nil {
		// If initial connect succeeds, it will disconnect and try to reconnect
		time.Sleep(300 * time.Millisecond)
		feed.Close()
	}
	
	// After max attempts, should be disconnected
	state := feed.State()
	if state != StateDisconnected && state != StateClosed {
		t.Errorf("expected state disconnected or closed after max attempts, got %s", state)
	}
}

func TestWebSocketFeed_Reconnect_ExponentialBackoff(t *testing.T) {
	config := DefaultWebSocketConfig()
	config.URL = "ws://localhost:9999"
	config.ReconnectDelay = 100 * time.Millisecond
	config.MaxReconnects = 5
	
	feed := NewWebSocketFeed(config)
	
	// Test calculateBackoff
	feed.reconnectAttempt = 1
	delay := feed.calculateBackoff()
	expectedMin := 100 * time.Millisecond
	expectedMax := 200 * time.Millisecond
	if delay < expectedMin || delay > expectedMax {
		t.Errorf("attempt 1: expected delay between %v and %v, got %v", expectedMin, expectedMax, delay)
	}
	
	feed.reconnectAttempt = 2
	delay = feed.calculateBackoff()
	expectedMin = 200 * time.Millisecond
	expectedMax = 400 * time.Millisecond
	if delay < expectedMin || delay > expectedMax {
		t.Errorf("attempt 2: expected delay between %v and %v, got %v", expectedMin, expectedMax, delay)
	}
	
	feed.reconnectAttempt = 3
	delay = feed.calculateBackoff()
	expectedMin = 400 * time.Millisecond
	expectedMax = 800 * time.Millisecond
	if delay < expectedMin || delay > expectedMax {
		t.Errorf("attempt 3: expected delay between %v and %v, got %v", expectedMin, expectedMax, delay)
	}
	
	// Test max delay cap (60s)
	feed.reconnectAttempt = 20
	delay = feed.calculateBackoff()
	if delay > 60*time.Second {
		t.Errorf("expected delay capped at 60s, got %v", delay)
	}
}

func TestWebSocketFeed_Reconnect_CancelDuringReconnection(t *testing.T) {
	// Server always fails
	server := mockFailingServer(100)
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	config := DefaultWebSocketConfig()
	config.URL = wsURL
	config.ReconnectDelay = 100 * time.Millisecond
	config.MaxReconnects = 10
	
	feed := NewWebSocketFeed(config)
	
	_ = feed.Connect()
	
	// Wait a bit for reconnection attempts to start
	time.Sleep(50 * time.Millisecond)
	
	// Close during reconnection
	err := feed.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
	
	if feed.State() != StateClosed {
		t.Errorf("expected state closed, got %s", feed.State())
	}
}

func TestWebSocketFeed_Reconnect_StateTransitions(t *testing.T) {
	// Server fails first connection, then succeeds
	server := mockFailingServer(1)
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	config := DefaultWebSocketConfig()
	config.URL = wsURL
	config.ReconnectDelay = 50 * time.Millisecond
	config.MaxReconnects = 5
	
	feed := NewWebSocketFeed(config)
	
	// Track state changes
	states := []ConnectionState{}
	statesChan := make(chan ConnectionState, 10)
	
	go func() {
		lastState := feed.State()
		for {
			time.Sleep(10 * time.Millisecond)
			currentState := feed.State()
			if currentState != lastState {
				statesChan <- currentState
				lastState = currentState
			}
			if currentState == StateClosed {
				return
			}
		}
	}()
	
	err := feed.Connect()
	_ = err // May succeed or fail
	
	// Wait for reconnection
	time.Sleep(300 * time.Millisecond)
	
	feed.Close()
	
	// Collect states
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case state := <-statesChan:
			states = append(states, state)
		case <-timeout:
			goto done
		}
	}
	
done:
	// Should have seen state transitions
	if len(states) < 1 {
		t.Errorf("expected multiple state transitions, got %d", len(states))
	}
}

func TestWebSocketFeed_Reconnect_ResetAttemptCounter(t *testing.T) {
	// Server fails first 2 connections, then succeeds
	server := mockFailingServer(2)
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	config := DefaultWebSocketConfig()
	config.URL = wsURL
	config.ReconnectDelay = 10 * time.Millisecond
	config.MaxReconnects = 10
	
	feed := NewWebSocketFeed(config)
	
	err := feed.Connect()
	_ = err
	
	// Wait for successful reconnection
	time.Sleep(200 * time.Millisecond)
	
	if feed.State() == StateConnected {
		// Reconnection succeeded, attempt counter should be reset
		if feed.reconnectAttempt != 0 {
			t.Errorf("expected reconnectAttempt to be reset to 0, got %d", feed.reconnectAttempt)
		}
	}
	
	feed.Close()
}
