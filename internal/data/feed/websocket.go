package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// ConnectionState represents the WebSocket connection state.
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateClosed
)

// String returns the string representation of the connection state.
func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateReconnecting:
		return "reconnecting"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// WebSocketFeed provides real-time market data via WebSocket.
type WebSocketFeed struct {
	url           string
	symbols       []string
	timeframe     time.Duration
	
	conn          *websocket.Conn
	state         ConnectionState
	stateMu       sync.RWMutex
	
	reconnectDelay   time.Duration
	maxReconnects    int
	reconnectAttempt int
	
	buffer        *CandleBuffer
	subscribers   []chan *market.Candle
	subMu         sync.RWMutex
	
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	
	lastPing      time.Time
	lastPingMu    sync.RWMutex
	pingInterval  time.Duration
	pongTimeout   time.Duration
	
	errChan       chan error
}

// WebSocketConfig holds WebSocket feed configuration.
type WebSocketConfig struct {
	URL              string
	Symbols          []string
	Timeframe        time.Duration
	ReconnectDelay   time.Duration
	MaxReconnects    int
	PingInterval     time.Duration
	PongTimeout      time.Duration
	BufferSize       int
}

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		ReconnectDelay: 1 * time.Second,
		MaxReconnects:  10,
		PingInterval:   30 * time.Second,
		PongTimeout:    10 * time.Second,
		BufferSize:     1000,
	}
}

// NewWebSocketFeed creates a new WebSocket feed.
func NewWebSocketFeed(config WebSocketConfig) *WebSocketFeed {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WebSocketFeed{
		url:              config.URL,
		symbols:          config.Symbols,
		timeframe:        config.Timeframe,
		state:            StateDisconnected,
		reconnectDelay:   config.ReconnectDelay,
		maxReconnects:    config.MaxReconnects,
		reconnectAttempt: 0,
		buffer:           NewCandleBuffer(config.BufferSize),
		subscribers:      make([]chan *market.Candle, 0),
		ctx:              ctx,
		cancel:           cancel,
		lastPing:         time.Now(),
		pingInterval:     config.PingInterval,
		pongTimeout:      config.PongTimeout,
		errChan:          make(chan error, 10),
	}
}

// Connect establishes the WebSocket connection.
func (f *WebSocketFeed) Connect() error {
	f.stateMu.Lock()
	if f.state == StateConnected || f.state == StateConnecting {
		f.stateMu.Unlock()
		return fmt.Errorf("already connected or connecting")
	}
	if f.state == StateClosed {
		f.stateMu.Unlock()
		return fmt.Errorf("feed is closed")
	}
	f.state = StateConnecting
	f.stateMu.Unlock()
	
	conn, _, err := websocket.DefaultDialer.Dial(f.url, nil)
	if err != nil {
		f.setState(StateDisconnected)
		return fmt.Errorf("dial failed: %w", err)
	}
	
	f.conn = conn
	f.setState(StateConnected)
	f.reconnectAttempt = 0
	
	// Start background goroutines
	f.wg.Add(3)
	go f.readLoop()
	go f.heartbeatLoop()
	go f.errorHandler()
	
	return nil
}

// Subscribe returns a channel that receives candle updates.
func (f *WebSocketFeed) Subscribe() <-chan *market.Candle {
	f.subMu.Lock()
	defer f.subMu.Unlock()
	
	ch := make(chan *market.Candle, 100)
	f.subscribers = append(f.subscribers, ch)
	return ch
}

// Close closes the WebSocket connection and all resources.
func (f *WebSocketFeed) Close() error {
	f.stateMu.Lock()
	if f.state == StateClosed {
		f.stateMu.Unlock()
		return nil
	}
	f.state = StateClosed
	f.stateMu.Unlock()
	
	// Cancel context to stop all goroutines
	f.cancel()
	
	// Close WebSocket connection
	if f.conn != nil {
		f.conn.Close()
	}
	
	// Wait for all goroutines to finish
	f.wg.Wait()
	
	// Close all subscriber channels
	f.subMu.Lock()
	for _, ch := range f.subscribers {
		close(ch)
	}
	f.subscribers = nil
	f.subMu.Unlock()
	
	close(f.errChan)
	
	return nil
}

// State returns the current connection state.
func (f *WebSocketFeed) State() ConnectionState {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return f.state
}

// setState updates the connection state.
func (f *WebSocketFeed) setState(state ConnectionState) {
	f.stateMu.Lock()
	f.state = state
	f.stateMu.Unlock()
}

// readLoop reads messages from the WebSocket connection.
func (f *WebSocketFeed) readLoop() {
	defer f.wg.Done()
	
	for {
		select {
		case <-f.ctx.Done():
			return
		default:
		}
		
		if f.conn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		
		_, message, err := f.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				f.errChan <- fmt.Errorf("read error: %w", err)
			}
			return
		}
		
		// Parse message into Candle
		candle, err := f.parseMessage(message)
		if err != nil {
			// Log parse error but don't disconnect
			continue
		}
		
		if candle != nil {
			// Buffer the candle
			err = f.buffer.Push(candle)
			if err != nil {
				// Buffer overflow, try to drain and retry
				drained := f.buffer.Drain()
				for _, c := range drained {
					f.broadcast(c)
				}
				f.buffer.Push(candle)
			}
			
			// Broadcast immediately
			f.broadcast(candle)
		}
		
		// Update last activity time
		f.lastPingMu.Lock()
		f.lastPing = time.Now()
		f.lastPingMu.Unlock()
	}
}

// parseMessage parses a WebSocket message into a Candle.
// This is a generic JSON parser. Exchange-specific implementations
// should override this in custom feed types.
func (f *WebSocketFeed) parseMessage(message []byte) (*market.Candle, error) {
	var data struct {
		Timestamp int64   `json:"timestamp"`
		Open      float64 `json:"open"`
		High      float64 `json:"high"`
		Low       float64 `json:"low"`
		Close     float64 `json:"close"`
		Volume    float64 `json:"volume"`
	}
	
	err := json.Unmarshal(message, &data)
	if err != nil {
		return nil, fmt.Errorf("parse message: %w", err)
	}
	
	candle := &market.Candle{
		Timestamp: time.Unix(data.Timestamp, 0),
		Open:      data.Open,
		High:      data.High,
		Low:       data.Low,
		Close:     data.Close,
		Volume:    data.Volume,
	}
	
	// Validate candle
	if !candle.IsValid() {
		return nil, fmt.Errorf("invalid candle: %+v", candle)
	}
	
	return candle, nil
}

// heartbeatLoop monitors connection health via ping/pong.
func (f *WebSocketFeed) heartbeatLoop() {
	defer f.wg.Done()
	
	ticker := time.NewTicker(f.pingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-f.ctx.Done():
			return
		case <-ticker.C:
			if f.State() != StateConnected {
				continue
			}
			
			// Check if we've received any message recently
			f.lastPingMu.RLock()
			lastActivity := f.lastPing
			f.lastPingMu.RUnlock()
			
			if time.Since(lastActivity) > f.pingInterval+f.pongTimeout {
				// Connection appears stale, trigger reconnection
				f.errChan <- fmt.Errorf("heartbeat timeout")
				return
			}
			
			// Send ping
			if f.conn != nil {
				err := f.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))
				if err != nil {
					f.errChan <- fmt.Errorf("ping failed: %w", err)
					return
				}
			}
		}
	}
}

// errorHandler handles errors and triggers reconnection.
func (f *WebSocketFeed) errorHandler() {
	defer f.wg.Done()
	
	for {
		select {
		case <-f.ctx.Done():
			return
		case err := <-f.errChan:
			if err == nil {
				continue
			}
			
			// Check if we should reconnect
			if f.State() == StateClosed {
				return
			}
			
			// Close current connection
			if f.conn != nil {
				f.conn.Close()
				f.conn = nil
			}
			
			// Trigger reconnection
			f.setState(StateDisconnected)
			f.reconnect()
		}
	}
}

// reconnect attempts to reconnect with exponential backoff.
func (f *WebSocketFeed) reconnect() {
	f.setState(StateReconnecting)
	
	for f.reconnectAttempt < f.maxReconnects {
		select {
		case <-f.ctx.Done():
			return
		default:
		}
		
		f.reconnectAttempt++
		
		// Calculate exponential backoff delay
		delay := f.calculateBackoff()
		
		time.Sleep(delay)
		
		// Attempt reconnection
		f.stateMu.Lock()
		if f.state == StateClosed {
			f.stateMu.Unlock()
			return
		}
		f.state = StateConnecting
		f.stateMu.Unlock()
		
		conn, _, err := websocket.DefaultDialer.Dial(f.url, nil)
		if err != nil {
			// Connection failed, continue loop
			f.setState(StateReconnecting)
			continue
		}
		
		// Reconnection successful
		f.conn = conn
		f.setState(StateConnected)
		f.reconnectAttempt = 0
		
		// Restart background goroutines
		f.wg.Add(2)
		go f.readLoop()
		go f.heartbeatLoop()
		
		return
	}
	
	// Max reconnection attempts reached
	f.setState(StateDisconnected)
}

// calculateBackoff calculates exponential backoff delay.
// Formula: min(baseDelay * 2^attempt, maxDelay)
func (f *WebSocketFeed) calculateBackoff() time.Duration {
	const maxDelay = 60 * time.Second
	
	delay := f.reconnectDelay * time.Duration(1<<uint(f.reconnectAttempt-1))
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return delay
}

// broadcast sends a candle to all subscribers.
func (f *WebSocketFeed) broadcast(candle *market.Candle) {
	f.subMu.RLock()
	defer f.subMu.RUnlock()
	
	for _, ch := range f.subscribers {
		select {
		case ch <- candle:
		default:
			// Subscriber channel full, skip
		}
	}
}
