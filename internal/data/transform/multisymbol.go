package transform

import (
	"context"
	"fmt"
	"sync"

	"github.com/ZulferDev/smallbt_go/internal/data"
	"github.com/ZulferDev/smallbt_go/internal/market"
)

// MultiSymbolTransformedFeed coordinates transforms across multiple symbols.
// Each symbol has its own transform chain and can be processed independently.
type MultiSymbolTransformedFeed struct {
	feeds map[string]*TransformedFeed
	mu    sync.RWMutex
}

// NewMultiSymbolTransformedFeed creates a multi-symbol transformed feed.
func NewMultiSymbolTransformedFeed() *MultiSymbolTransformedFeed {
	return &MultiSymbolTransformedFeed{
		feeds: make(map[string]*TransformedFeed),
	}
}

// AddSymbol adds a symbol with its own feed and transform chain.
func (msf *MultiSymbolTransformedFeed) AddSymbol(symbol string, feed data.DataFeed, chain *TransformChain, batchSize int) error {
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	msf.mu.Lock()
	defer msf.mu.Unlock()

	if _, exists := msf.feeds[symbol]; exists {
		return fmt.Errorf("symbol %s already exists", symbol)
	}

	tfeed, err := NewTransformedFeed(feed, chain, batchSize)
	if err != nil {
		return fmt.Errorf("create transformed feed for %s: %w", symbol, err)
	}

	msf.feeds[symbol] = tfeed
	return nil
}

// RemoveSymbol removes a symbol and closes its feed.
func (msf *MultiSymbolTransformedFeed) RemoveSymbol(symbol string) error {
	msf.mu.Lock()
	defer msf.mu.Unlock()

	tfeed, exists := msf.feeds[symbol]
	if !exists {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	if err := tfeed.Close(); err != nil {
		return fmt.Errorf("close feed for %s: %w", symbol, err)
	}

	delete(msf.feeds, symbol)
	return nil
}

// Next returns the next candle for a specific symbol.
func (msf *MultiSymbolTransformedFeed) Next(symbol string) (*market.Candle, error) {
	msf.mu.RLock()
	tfeed, exists := msf.feeds[symbol]
	msf.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	return tfeed.Next()
}

// NextAll returns the next candle for all symbols.
// Returns a map of symbol -> candle. Symbols with errors are omitted.
func (msf *MultiSymbolTransformedFeed) NextAll() (map[string]*market.Candle, error) {
	msf.mu.RLock()
	defer msf.mu.RUnlock()

	if len(msf.feeds) == 0 {
		return nil, fmt.Errorf("no symbols configured")
	}

	result := make(map[string]*market.Candle)
	var errs []error

	for symbol, tfeed := range msf.feeds {
		candle, err := tfeed.Next()
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", symbol, err))
			continue
		}
		result[symbol] = candle
	}

	// If all symbols failed, return error
	if len(result) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all symbols failed: %v", errs)
	}

	return result, nil
}

// ReadAll reads all data for a specific symbol.
func (msf *MultiSymbolTransformedFeed) ReadAll(symbol string) ([]*market.Candle, error) {
	msf.mu.RLock()
	tfeed, exists := msf.feeds[symbol]
	msf.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	return tfeed.ReadAll()
}

// ReadAllSymbols reads all data for all symbols.
// Returns a map of symbol -> candles.
func (msf *MultiSymbolTransformedFeed) ReadAllSymbols() (map[string][]*market.Candle, error) {
	msf.mu.RLock()
	defer msf.mu.RUnlock()

	if len(msf.feeds) == 0 {
		return nil, fmt.Errorf("no symbols configured")
	}

	result := make(map[string][]*market.Candle)
	var errs []error

	for symbol, tfeed := range msf.feeds {
		candles, err := tfeed.ReadAll()
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", symbol, err))
			continue
		}
		result[symbol] = candles
	}

	if len(result) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all symbols failed: %v", errs)
	}

	return result, nil
}

// Symbols returns the list of configured symbols.
func (msf *MultiSymbolTransformedFeed) Symbols() []string {
	msf.mu.RLock()
	defer msf.mu.RUnlock()

	symbols := make([]string, 0, len(msf.feeds))
	for symbol := range msf.feeds {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// Close closes all feeds.
func (msf *MultiSymbolTransformedFeed) Close() error {
	msf.mu.Lock()
	defer msf.mu.Unlock()

	var errs []error
	for symbol, tfeed := range msf.feeds {
		if err := tfeed.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", symbol, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// ParallelMultiSymbolFeed processes multiple symbols in parallel.
type ParallelMultiSymbolFeed struct {
	feeds   map[string]*TransformedFeed
	mu      sync.RWMutex
	workers int
}

// NewParallelMultiSymbolFeed creates a parallel multi-symbol feed.
func NewParallelMultiSymbolFeed(workers int) *ParallelMultiSymbolFeed {
	if workers <= 0 {
		workers = 4 // Default
	}
	return &ParallelMultiSymbolFeed{
		feeds:   make(map[string]*TransformedFeed),
		workers: workers,
	}
}

// AddSymbol adds a symbol.
func (pmsf *ParallelMultiSymbolFeed) AddSymbol(symbol string, feed data.DataFeed, chain *TransformChain, batchSize int) error {
	pmsf.mu.Lock()
	defer pmsf.mu.Unlock()

	if _, exists := pmsf.feeds[symbol]; exists {
		return fmt.Errorf("symbol %s already exists", symbol)
	}

	tfeed, err := NewTransformedFeed(feed, chain, batchSize)
	if err != nil {
		return fmt.Errorf("create transformed feed for %s: %w", symbol, err)
	}

	pmsf.feeds[symbol] = tfeed
	return nil
}

// ReadAllParallel reads all symbols in parallel.
func (pmsf *ParallelMultiSymbolFeed) ReadAllParallel(ctx context.Context) (map[string][]*market.Candle, error) {
	pmsf.mu.RLock()
	symbols := make([]string, 0, len(pmsf.feeds))
	for symbol := range pmsf.feeds {
		symbols = append(symbols, symbol)
	}
	pmsf.mu.RUnlock()

	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols configured")
	}

	// Result channel
	type result struct {
		symbol  string
		candles []*market.Candle
		err     error
	}
	resultCh := make(chan result, len(symbols))

	// Worker pool
	symbolCh := make(chan string, len(symbols))
	for _, symbol := range symbols {
		symbolCh <- symbol
	}
	close(symbolCh)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < pmsf.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for symbol := range symbolCh {
				select {
				case <-ctx.Done():
					resultCh <- result{symbol: symbol, err: ctx.Err()}
					return
				default:
				}

				pmsf.mu.RLock()
				tfeed := pmsf.feeds[symbol]
				pmsf.mu.RUnlock()

				candles, err := tfeed.ReadAll()
				resultCh <- result{symbol: symbol, candles: candles, err: err}
			}
		}()
	}

	// Wait and close
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	results := make(map[string][]*market.Candle)
	var errs []error

	for res := range resultCh {
		if res.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", res.symbol, res.err))
			continue
		}
		results[res.symbol] = res.candles
	}

	if len(results) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all symbols failed: %v", errs)
	}

	return results, nil
}

// Close closes all feeds.
func (pmsf *ParallelMultiSymbolFeed) Close() error {
	pmsf.mu.Lock()
	defer pmsf.mu.Unlock()

	var errs []error
	for symbol, tfeed := range pmsf.feeds {
		if err := tfeed.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", symbol, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}
