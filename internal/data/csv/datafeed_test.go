package csv

import (
	"context"
	"io"
	"testing"

	"github.com/ZulferDev/smallbt_go/internal/data"
)

func TestCSVDataFeed_Interface(t *testing.T) {
	// Verify CSVDataFeed implements data.DataFeed
	var _ data.DataFeed = (*CSVDataFeed)(nil)
}

func TestCSVDataFeed_Subscribe(t *testing.T) {
	feed, err := NewCSVDataFeed("../../../data/test/BTCUSDT_sample.csv",
		DefaultCSVConfig("BTCUSDT", "1h"))
	if err != nil {
		t.Skip("test data not available")
	}
	defer feed.Close()

	ctx := context.Background()

	// Subscribe with matching symbol
	err = feed.Subscribe(ctx, []string{"BTCUSDT"})
	if err != nil {
		t.Errorf("Subscribe() with matching symbol failed: %v", err)
	}

	// Subscribe with empty list (should succeed)
	err = feed.Subscribe(ctx, []string{})
	if err != nil {
		t.Errorf("Subscribe() with empty list failed: %v", err)
	}

	// Subscribe with wrong symbol
	err = feed.Subscribe(ctx, []string{"ETHUSDT"})
	if err != data.ErrSymbolNotSupported {
		t.Errorf("Subscribe() with wrong symbol: got %v, want %v", err, data.ErrSymbolNotSupported)
	}
}

func TestCSVDataFeed_Next(t *testing.T) {
	feed, err := NewCSVDataFeed("../../../data/test/BTCUSDT_sample.csv",
		DefaultCSVConfig("BTCUSDT", "1h"))
	if err != nil {
		t.Skip("test data not available")
	}
	defer feed.Close()

	ctx := context.Background()

	// Read first candle
	candle, err := feed.Next(ctx)
	if err != nil {
		t.Fatalf("Next() failed: %v", err)
	}
	if candle == nil {
		t.Fatal("Next() returned nil candle")
	}

	// Candle should have valid OHLC
	if !candle.IsValid() {
		t.Error("Next() returned invalid candle")
	}
}

func TestCSVDataFeed_EOF(t *testing.T) {
	// Create a minimal CSV for testing
	feed, err := NewCSVDataFeed("../../../data/test/BTCUSDT_sample.csv",
		DefaultCSVConfig("BTCUSDT", "1h"))
	if err != nil {
		t.Skip("test data not available")
	}
	defer feed.Close()

	ctx := context.Background()

	// Read all candles
	count := 0
	for {
		_, err := feed.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next() failed: %v", err)
		}
		count++
		if count > 10000 {
			t.Fatal("read too many candles, expected EOF")
		}
	}

	// Next call should still return EOF
	_, err = feed.Next(ctx)
	if err != io.EOF {
		t.Errorf("After EOF, Next() returned %v, want io.EOF", err)
	}
}

func TestCSVDataFeed_Close(t *testing.T) {
	feed, err := NewCSVDataFeed("../../../data/test/BTCUSDT_sample.csv",
		DefaultCSVConfig("BTCUSDT", "1h"))
	if err != nil {
		t.Skip("test data not available")
	}

	ctx := context.Background()

	// Close feed
	err = feed.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Operations after Close should fail
	err = feed.Subscribe(ctx, []string{"BTCUSDT"})
	if err != data.ErrFeedClosed {
		t.Errorf("Subscribe() after Close: got %v, want %v", err, data.ErrFeedClosed)
	}

	_, err = feed.Next(ctx)
	if err != data.ErrFeedClosed {
		t.Errorf("Next() after Close: got %v, want %v", err, data.ErrFeedClosed)
	}
}

func TestCSVDataFeed_ContextCancellation(t *testing.T) {
	feed, err := NewCSVDataFeed("../../../data/test/BTCUSDT_sample.csv",
		DefaultCSVConfig("BTCUSDT", "1h"))
	if err != nil {
		t.Skip("test data not available")
	}
	defer feed.Close()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Next should respect context cancellation
	_, err = feed.Next(ctx)
	if err != context.Canceled {
		t.Errorf("Next() with cancelled context: got %v, want %v", err, context.Canceled)
	}
}

func TestCSVDataFeed_Reset(t *testing.T) {
	feed, err := NewCSVDataFeed("../../../data/test/BTCUSDT_sample.csv",
		DefaultCSVConfig("BTCUSDT", "1h"))
	if err != nil {
		t.Skip("test data not available")
	}
	defer feed.Close()

	ctx := context.Background()

	// Read first candle
	first, err := feed.Next(ctx)
	if err != nil {
		t.Fatalf("Next() failed: %v", err)
	}

	// Read second candle
	_, err = feed.Next(ctx)
	if err != nil {
		t.Fatalf("Next() failed: %v", err)
	}

	// Reset
	feed.Reset()

	// Should read first candle again
	firstAgain, err := feed.Next(ctx)
	if err != nil {
		t.Fatalf("Next() after Reset failed: %v", err)
	}

	if firstAgain.Timestamp != first.Timestamp {
		t.Errorf("After Reset, first candle timestamp mismatch: got %v, want %v",
			firstAgain.Timestamp, first.Timestamp)
	}
}
