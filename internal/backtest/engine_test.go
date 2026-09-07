package backtest

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/execution"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCandles(t *testing.T) {
	// Create temporary test data directory
	tmpDir := t.TempDir()

	t.Run("load CSV file", func(t *testing.T) {
		// Create test CSV file
		csvPath := filepath.Join(tmpDir, "test.csv")
		csvContent := `timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,100,105,95,102,1000
2024-01-01T01:00:00Z,102,108,100,105,1200
2024-01-01T02:00:00Z,105,110,103,108,1500
`
		err := os.WriteFile(csvPath, []byte(csvContent), 0644)
		require.NoError(t, err)

		// Load candles
		candles, err := loadCandles(csvPath, market.Timeframe1h)
		require.NoError(t, err)
		assert.NotEmpty(t, candles)

		// Verify first candle
		assert.Equal(t, 100.0, candles[0].Open)
		assert.Equal(t, 105.0, candles[0].High)
		assert.Equal(t, 95.0, candles[0].Low)
		assert.Equal(t, 102.0, candles[0].Close)
		assert.Equal(t, 1000.0, candles[0].Volume)
	})

	t.Run("unsupported file format", func(t *testing.T) {
		txtPath := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(txtPath, []byte("invalid"), 0644)
		require.NoError(t, err)

		_, err = loadCandles(txtPath, market.Timeframe1h)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported file format")
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := loadCandles(filepath.Join(tmpDir, "nonexistent.csv"), market.Timeframe1h)
		assert.Error(t, err)
	})

	t.Run("relative path", func(t *testing.T) {
		// Create CSV in temp dir
		csvPath := filepath.Join(tmpDir, "relative.csv")
		csvContent := `timestamp,open,high,low,close,volume
2024-01-01T00:00:00Z,100,100,100,100,100
`
		err := os.WriteFile(csvPath, []byte(csvContent), 0644)
		require.NoError(t, err)

		// Load with absolute path (function converts relative to absolute)
		candles, err := loadCandles(csvPath, market.Timeframe1h)
		require.NoError(t, err)
		assert.Len(t, candles, 1)
	})
}

func TestFilterCandlesByTime(t *testing.T) {
	// Create test candles
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := []*market.Candle{
		{Timestamp: baseTime, Close: 100},
		{Timestamp: baseTime.Add(1 * time.Hour), Close: 101},
		{Timestamp: baseTime.Add(2 * time.Hour), Close: 102},
		{Timestamp: baseTime.Add(3 * time.Hour), Close: 103},
		{Timestamp: baseTime.Add(4 * time.Hour), Close: 104},
	}

	t.Run("filter with start time only", func(t *testing.T) {
		start := baseTime.Add(2 * time.Hour)
		filtered := filterCandlesByTime(candles, start, time.Time{})

		assert.Len(t, filtered, 3)
		assert.Equal(t, 102.0, filtered[0].Close)
		assert.Equal(t, 103.0, filtered[1].Close)
		assert.Equal(t, 104.0, filtered[2].Close)
	})

	t.Run("filter with end time only", func(t *testing.T) {
		end := baseTime.Add(3 * time.Hour)
		filtered := filterCandlesByTime(candles, time.Time{}, end)

		assert.Len(t, filtered, 3)
		assert.Equal(t, 100.0, filtered[0].Close)
		assert.Equal(t, 101.0, filtered[1].Close)
		assert.Equal(t, 102.0, filtered[2].Close)
	})

	t.Run("filter with both start and end", func(t *testing.T) {
		start := baseTime.Add(1 * time.Hour)
		end := baseTime.Add(3 * time.Hour)
		filtered := filterCandlesByTime(candles, start, end)

		assert.Len(t, filtered, 2)
		assert.Equal(t, 101.0, filtered[0].Close)
		assert.Equal(t, 102.0, filtered[1].Close)
	})

	t.Run("no time filters", func(t *testing.T) {
		filtered := filterCandlesByTime(candles, time.Time{}, time.Time{})

		assert.Len(t, filtered, 5)
		assert.Equal(t, candles, filtered)
	})

	t.Run("start equals end", func(t *testing.T) {
		start := baseTime.Add(2 * time.Hour)
		end := baseTime.Add(2 * time.Hour)
		filtered := filterCandlesByTime(candles, start, end)

		assert.Empty(t, filtered)
	})

	t.Run("start after all candles", func(t *testing.T) {
		start := baseTime.Add(10 * time.Hour)
		filtered := filterCandlesByTime(candles, start, time.Time{})

		assert.Empty(t, filtered)
	})

	t.Run("end before all candles", func(t *testing.T) {
		end := baseTime.Add(-1 * time.Hour)
		filtered := filterCandlesByTime(candles, time.Time{}, end)

		assert.Empty(t, filtered)
	})

	t.Run("empty candle slice", func(t *testing.T) {
		filtered := filterCandlesByTime([]*market.Candle{}, baseTime, baseTime.Add(1*time.Hour))

		assert.Empty(t, filtered)
	})
}

func TestCreateSlippageModel(t *testing.T) {
	// Helper to create test order request and candle
	createTestOrder := func() order.OrderRequest {
		return order.OrderRequest{
			Symbol:   market.Symbol("BTCUSDT"),
			Side:     order.OrderSideBuy,
			Type:     order.OrderTypeMarket,
			Quantity: 1.0,
		}
	}

	createTestCandle := func() *market.Candle {
		return &market.Candle{
			Timestamp: time.Now(),
			Open:      100.0,
			High:      105.0,
			Low:       95.0,
			Close:     102.0,
			Volume:    1000.0,
		}
	}

	t.Run("fixed slippage model", func(t *testing.T) {
		params := map[string]float64{"value": 0.5}
		model := createSlippageModel("fixed", params)

		assert.NotNil(t, model)
		req := createTestOrder()
		candle := createTestCandle()
		slippage, err := model.CalculateSlippage(req, 100.0, candle)
		require.NoError(t, err)
		assert.Equal(t, 0.5, slippage)
	})

	t.Run("percentage slippage model", func(t *testing.T) {
		params := map[string]float64{"value": 0.001}
		model := createSlippageModel("percentage", params)

		assert.NotNil(t, model)
		req := createTestOrder()
		candle := createTestCandle()
		slippage, err := model.CalculateSlippage(req, 100.0, candle)
		require.NoError(t, err)
		assert.Equal(t, 0.1, slippage) // 0.1% of 100
	})

	t.Run("volatility slippage model", func(t *testing.T) {
		params := map[string]float64{"value": 0.5, "max": 2.0}
		model := createSlippageModel("volatility", params)

		assert.NotNil(t, model)
		req := createTestOrder()
		candle := createTestCandle()
		slippage, err := model.CalculateSlippage(req, 100.0, candle)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, slippage, 0.0)
		assert.LessOrEqual(t, slippage, 2.0) // Respects max
	})

	t.Run("volume slippage model", func(t *testing.T) {
		params := map[string]float64{"value": 0.01, "max": 1.0}
		model := createSlippageModel("volume", params)

		assert.NotNil(t, model)
		req := createTestOrder()
		candle := createTestCandle()
		slippage, err := model.CalculateSlippage(req, 100.0, candle)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, slippage, 0.0)
		assert.LessOrEqual(t, slippage, 1.0)
	})

	t.Run("no slippage model", func(t *testing.T) {
		params := map[string]float64{}
		model := createSlippageModel("none", params)

		assert.NotNil(t, model)
		req := createTestOrder()
		candle := createTestCandle()
		slippage, err := model.CalculateSlippage(req, 100.0, candle)
		require.NoError(t, err)
		assert.Equal(t, 0.0, slippage)
	})

	t.Run("unknown model defaults to no slippage", func(t *testing.T) {
		params := map[string]float64{"value": 1.0}
		model := createSlippageModel("unknown_model", params)

		assert.NotNil(t, model)
		req := createTestOrder()
		candle := createTestCandle()
		slippage, err := model.CalculateSlippage(req, 100.0, candle)
		require.NoError(t, err)
		assert.Equal(t, 0.0, slippage)
	})

	t.Run("empty model name", func(t *testing.T) {
		params := map[string]float64{}
		model := createSlippageModel("", params)

		assert.NotNil(t, model)
		req := createTestOrder()
		candle := createTestCandle()
		slippage, err := model.CalculateSlippage(req, 100.0, candle)
		require.NoError(t, err)
		assert.Equal(t, 0.0, slippage)
	})

	t.Run("missing params uses zero values", func(t *testing.T) {
		// Empty params map
		model := createSlippageModel("fixed", map[string]float64{})

		assert.NotNil(t, model)
		req := createTestOrder()
		candle := createTestCandle()
		slippage, err := model.CalculateSlippage(req, 100.0, candle)
		require.NoError(t, err)
		assert.Equal(t, 0.0, slippage) // value defaults to 0
	})
}

func TestCreateSlippageModelTypes(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		params    map[string]float64
	}{
		{
			name:      "fixed",
			modelName: "fixed",
			params:    map[string]float64{"value": 0.1},
		},
		{
			name:      "percentage",
			modelName: "percentage",
			params:    map[string]float64{"value": 0.001},
		},
		{
			name:      "volatility",
			modelName: "volatility",
			params:    map[string]float64{"value": 0.5, "max": 2.0},
		},
		{
			name:      "volume",
			modelName: "volume",
			params:    map[string]float64{"value": 0.01, "max": 1.0},
		},
		{
			name:      "none",
			modelName: "none",
			params:    map[string]float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := createSlippageModel(tt.modelName, tt.params)
			assert.NotNil(t, model)

			// Verify model implements execution.SlippageModel interface
			var _ execution.SlippageModel = model
		})
	}
}
