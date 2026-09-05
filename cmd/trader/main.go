package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/ZulferDev/smallbt_go/internal/backtest"
	"github.com/ZulferDev/smallbt_go/internal/broker"
	"github.com/ZulferDev/smallbt_go/internal/data/csv"
	"github.com/ZulferDev/smallbt_go/internal/execution"
	"github.com/ZulferDev/smallbt_go/internal/market"
	"github.com/ZulferDev/smallbt_go/internal/montecarlo"
	"github.com/ZulferDev/smallbt_go/internal/optimization"
	"github.com/ZulferDev/smallbt_go/internal/portfolio"
	"github.com/ZulferDev/smallbt_go/internal/strategy/parser"
	"github.com/ZulferDev/smallbt_go/internal/walkforward"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("expected command, try 'validate', 'backtest', 'optimize', 'walkforward', or 'report'")
	}

	command := os.Args[1]

	switch command {
	case "validate":
		return runValidate(os.Args[2:])
	case "backtest":
		return runBacktest(os.Args[2:])
	case "optimize":
		return runOptimize(os.Args[2:])
	case "walkforward":
		return runWalkforward(os.Args[2:])
	case "montecarlo":
		return runMonteCarlo(os.Args[2:])
	case "paper":
		return runPaper(os.Args[2:])
	case "report":
		return runReport(os.Args[2:])
	case "-h", "--help", "help":
		printHelp()
		return nil
	case "-v", "--version":
		fmt.Println("trader v0.1.0-alpha (Phase 0 - Architecture Foundation)")
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printHelp() {
	fmt.Println(`trader - Declarative Quantitative Trading Backtesting Engine

COMMANDS:
  validate      Validate a strategy YAML configuration
  backtest      Run a backtest with a strategy and data
  optimize      Optimize strategy parameters
  walkforward   Run Walk Forward Analysis
  montecarlo    Run Monte Carlo Simulation
  paper         Run paper trading with simulated real-time data
  report        Generate reports from backtest results

FLAGS:
  -h, --help    Show this help message
  -v, --version Show version information

EXAMPLES:
  trader validate strategy.yaml
  trader backtest --strategy strategy.yaml --data data.csv
  trader paper --strategy strategy.yaml --symbol BTCUSDT --price 50000
  trader montecarlo --result backtest_result.json --simulations 10000`)
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	strategyPath := fs.String("strategy", "", "Path to strategy YAML file")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if *strategyPath == "" {
		// Try to use first positional argument
		if len(fs.Args()) > 0 {
			*strategyPath = fs.Arg(0)
		}
	}

	if *strategyPath == "" {
		return fmt.Errorf("strategy file path required")
	}

	// Parse and validate strategy
	parser := parser.NewParser()
	strategy, err := parser.ParseFile(*strategyPath)
	if err != nil {
		return fmt.Errorf("validate strategy: %w", err)
	}

	// Basic validation
	if strategy.Name == "" {
		fmt.Println("⚠️  Warning: Strategy name is empty")
	}

	if strategy.Data.Symbol == "" {
		fmt.Println("⚠️  Warning: Data symbol not specified")
	}

	if len(strategy.Indicators) == 0 {
		fmt.Println("⚠️  Warning: No indicators defined")
	}

	if strategy.Entry.Long == nil && strategy.Entry.Short == nil {
		fmt.Println("⚠️  Warning: No entry rules defined (long or short)")
	}

	fmt.Printf("✅ Strategy '%s' (v%s) validated successfully\n", strategy.Name, strategy.Version)
	fmt.Printf("   Symbol: %s, Timeframe: %s\n", strategy.Data.Symbol, strategy.Data.Timeframe)
	fmt.Printf("   Indicators: %d\n", len(strategy.Indicators))

	// Count entry rules
	entryRules := 0
	if strategy.Entry.Long != nil {
		entryRules++
	}
	if strategy.Entry.Short != nil {
		entryRules++
	}
	fmt.Printf("   Entry rules: %d\n", entryRules)

	// Show indicators
	for name, indicator := range strategy.Indicators {
		fmt.Printf("   - %s: %s (period: %d)\n", name, indicator.Type, indicator.Period)
	}

	return nil
}

func runBacktest(args []string) error {
	fs := flag.NewFlagSet("backtest", flag.ExitOnError)
	strategyPath := fs.String("strategy", "", "Path to strategy YAML file")
	dataPath := fs.String("data", "", "Path to data file (CSV/Parquet)")
	initialCash := fs.Float64("cash", 10000.0, "Initial cash amount")
	outputJSON := fs.String("output", "backtest_result.json", "Output JSON file for results")
	symbol := fs.String("symbol", "", "Symbol (e.g., BTCUSDT)")
	timeframe := fs.String("timeframe", "1h", "Timeframe (e.g., 1m, 5m, 15m, 30m, 1h, 4h, 1d)")
	startDate := fs.String("start", "", "Start date (YYYY-MM-DD)")
	endDate := fs.String("end", "", "End date (YYYY-MM-DD)")
	cpuProfile := fs.String("cpuprofile", "", "Write CPU profile to file")
	memProfile := fs.String("memprofile", "", "Write memory profile to file")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	// Start CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			return fmt.Errorf("create CPU profile: %w", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("start CPU profile: %w", err)
		}
		defer pprof.StopCPUProfile()
		fmt.Printf("CPU profiling enabled: %s\n", *cpuProfile)
	}

	// Validate required flags
	if *strategyPath == "" {
		return fmt.Errorf("strategy file required (--strategy)")
	}
	if *dataPath == "" {
		return fmt.Errorf("data file required (--data)")
	}
	if *symbol == "" {
		// Try to get from strategy first
		parser := parser.NewParser()
		strategy, err := parser.ParseFile(*strategyPath)
		if err == nil && strategy.Data.Symbol != "" {
			*symbol = strategy.Data.Symbol
		}
	}
	if *symbol == "" {
		return fmt.Errorf("symbol required (--symbol or in strategy)")
	}

	// Create config
	config := backtest.BacktestConfig{
		Symbol:       market.Symbol(*symbol),
		Timeframe:    market.Timeframe(*timeframe),
		InitialCash:  *initialCash,
		StrategyPath: *strategyPath,
		DataPath:     *dataPath,
	}

	// Parse dates if provided
	if *startDate != "" {
		startTime, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			return fmt.Errorf("invalid start date format (use YYYY-MM-DD): %w", err)
		}
		config.StartTime = startTime
	}
	if *endDate != "" {
		endTime, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			return fmt.Errorf("invalid end date format (use YYYY-MM-DD): %w", err)
		}
		config.EndTime = endTime
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Running backtest...")
	fmt.Printf("Strategy: %s\n", *strategyPath)
	fmt.Printf("Data:     %s\n", *dataPath)
	fmt.Printf("Symbol:   %s\n", *symbol)
	fmt.Printf("Cash:     $%.2f\n", *initialCash)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Run backtest
	start := time.Now()
	result, err := backtest.Run(config)
	if err != nil {
		return fmt.Errorf("backtest execution failed: %w", err)
	}
	elapsed := time.Since(start)

	// Display results
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("BACKTEST RESULT")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Strategy       %s\n", result.StrategyName)
	fmt.Printf("Symbol         %s\n", result.Config.Symbol)
	fmt.Printf("Timeframe      %s\n", result.Config.Timeframe)
	fmt.Printf("Period         %s → %s\n", result.StartTime.Format("2006-01-02"), result.EndTime.Format("2006-01-02"))
	fmt.Printf("Runtime        %v\n", elapsed)
	fmt.Println()

	fmt.Printf("Return         %+.2f%%\n", result.Metrics.TotalReturn*100)
	fmt.Printf("CAGR           %+.2f%%\n", result.Metrics.CAGR*100)
	fmt.Printf("Sharpe         %.2f\n", result.Metrics.SharpeRatio)
	fmt.Printf("Sortino        %.2f\n", result.Metrics.SortinoRatio)
	fmt.Printf("Max Drawdown   %.2f%%\n", result.Metrics.MaxDrawdown*100)
	fmt.Println()

	fmt.Printf("Trades         %d\n", result.TotalTrades)
	fmt.Printf("Win Rate       %.2f%%\n", result.Metrics.WinRate*100)
	fmt.Printf("Profit Factor  %.2f\n", result.Metrics.ProfitFactor)
	fmt.Printf("Expectancy     %.2fR\n", result.Metrics.Expectancy)
	fmt.Println()

	fmt.Printf("Final Equity   $%.2f\n", result.Portfolio.Equity)
	fmt.Printf("Total Fees     $%.2f\n", result.Metrics.TotalFees)
	fmt.Printf("Net PnL        $%.2f\n", result.Portfolio.Equity-result.Config.InitialCash)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Save to JSON
	if *outputJSON != "" {
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal results to JSON: %w", err)
		}
		if err := os.WriteFile(*outputJSON, jsonBytes, 0644); err != nil {
			return fmt.Errorf("write JSON file: %w", err)
		}
		fmt.Printf("Results saved to %s\n", *outputJSON)
	}

	// Write memory profile if requested
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			return fmt.Errorf("create memory profile: %w", err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			return fmt.Errorf("write memory profile: %w", err)
		}
		fmt.Printf("Memory profile saved to %s\n", *memProfile)
	}

	return nil
}

func runOptimize(args []string) error {
	fs := flag.NewFlagSet("optimize", flag.ExitOnError)
	strategyPath := fs.String("strategy", "", "Path to strategy YAML file")
	dataPath := fs.String("data", "", "Path to market data file")
	symbol := fs.String("symbol", "BTCUSDT", "Trading symbol")
	initialCash := fs.Float64("cash", 10000, "Initial cash")
	startDate := fs.String("start", "", "Start date (YYYY-MM-DD)")
	endDate := fs.String("end", "", "End date (YYYY-MM-DD)")
	outputJSON := fs.String("output", "", "Output JSON file path")
	parameters := fs.String("parameters", "", "Parameter ranges (format: name:start:end:step,name2:...)")
	objective := fs.String("objective", "sharpe", "Optimization objective (sharpe, sortino, return, profit_factor)")
	direction := fs.String("direction", "maximize", "Optimization direction (maximize or minimize)")
	parallel := fs.Int("parallel", 1, "Number of parallel workers (1 = sequential)")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if *strategyPath == "" {
		return fmt.Errorf("--strategy flag required")
	}
	if *dataPath == "" {
		return fmt.Errorf("--data flag required")
	}
	if *parameters == "" {
		return fmt.Errorf("--parameters flag required (format: name:start:end:step)")
	}

	// Parse parameter ranges
	paramRanges, err := parseParameterRanges(*parameters)
	if err != nil {
		return fmt.Errorf("parse parameter ranges: %w", err)
	}

	// Parse strategy to get timeframe
	p := parser.NewParser()
	strategyAST, err := p.ParseFile(*strategyPath)
	if err != nil {
		return fmt.Errorf("parse strategy: %w", err)
	}

	// Create optimization config
	config := optimization.OptimizationConfig{
		StrategyPath: *strategyPath,
		BacktestConfig: backtest.BacktestConfig{
			Symbol:       market.Symbol(*symbol),
			Timeframe:    market.Timeframe(strategyAST.Data.Timeframe),
			InitialCash:  *initialCash,
			StartTime:    time.Time{},
			EndTime:      time.Now(),
			StrategyPath: *strategyPath,
			DataPath:     *dataPath,
		},
		Parameters: paramRanges,
		Objective: optimization.ObjectiveConfig{
			Type:      *objective,
			Direction: *direction,
		},
		Algorithm: "grid",
	}

	// Parse dates
	if *startDate != "" {
		startTime, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			return fmt.Errorf("invalid start date: %w", err)
		}
		config.BacktestConfig.StartTime = startTime
	}
	if *endDate != "" {
		endTime, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			return fmt.Errorf("invalid end date: %w", err)
		}
		config.BacktestConfig.EndTime = endTime
	}

	// Display optimization info
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("PARAMETER OPTIMIZATION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Strategy:    %s\n", *strategyPath)
	fmt.Printf("Data:        %s\n", *dataPath)
	fmt.Printf("Symbol:      %s\n", *symbol)
	fmt.Printf("Objective:   %s (%s)\n", *objective, *direction)
	fmt.Printf("Algorithm:   Grid Search\n")
	fmt.Printf("Parameters:  %d\n", len(paramRanges))
	for _, p := range paramRanges {
		fmt.Printf("  - %s: [%.2f to %.2f, step %.2f]\n", p.Name, p.Start, p.End, p.Step)
	}

	// Create optimizer
	gridSearch := optimization.NewGridSearch(config)

	// Estimate total combinations
	totalCombinations := gridSearch.EstimateTotalCombinations()
	fmt.Printf("\nTotal Combinations: %d\n", totalCombinations)
	fmt.Printf("Parallel Workers:   %d\n", *parallel)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Run optimization
	start := time.Now()

	evaluator := func(paramSet optimization.ParameterSet) (*backtest.BacktestResult, error) {
		// Create modified strategy for this parameter set
		modifier := optimization.NewYAMLModifier()
		modifiedPath, err := modifier.ModifyStrategyFile(*strategyPath, paramSet)
		if err != nil {
			return nil, fmt.Errorf("modify strategy: %w", err)
		}
		defer os.Remove(modifiedPath) // Clean up temp file

		// Run backtest with modified strategy
		btConfig := backtest.BacktestConfig{
			Symbol:       config.BacktestConfig.Symbol,
			Timeframe:    config.BacktestConfig.Timeframe,
			StartTime:    config.BacktestConfig.StartTime,
			EndTime:      config.BacktestConfig.EndTime,
			InitialCash:  config.BacktestConfig.InitialCash,
			StrategyPath: modifiedPath,
			DataPath:     config.BacktestConfig.DataPath,
		}

		return backtest.Run(btConfig)
	}

	report, err := gridSearch.Run(evaluator, *parallel)
	if err != nil {
		return fmt.Errorf("optimization failed: %w", err)
	}

	elapsed := time.Since(start)

	// Display results
	fmt.Println(report.GenerateReport())
	fmt.Printf("\nOptimization completed in %v\n", elapsed)

	// Save to JSON
	if *outputJSON != "" {
		if err := report.SaveJSON(*outputJSON); err != nil {
			return fmt.Errorf("save JSON: %w", err)
		}
		fmt.Printf("Results saved to %s\n", *outputJSON)
	}

	return nil
}

// parseParameterRanges parses parameter range string.
// Format: "name1:start1:end1:step1,name2:start2:end2:step2"
// Example: "indicators.ema_fast.period:5:20:1,indicators.ema_slow.period:20:100:5"
// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseParameterRanges(input string) ([]optimization.ParameterRange, error) {
	var ranges []optimization.ParameterRange

	parts := strings.Split(input, ",")
	for _, part := range parts {
		fields := strings.Split(part, ":")
		if len(fields) != 4 {
			return nil, fmt.Errorf("invalid parameter format '%s' (expected name:start:end:step)", part)
		}

		name := fields[0]
		start, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid start value for %s: %w", name, err)
		}
		end, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid end value for %s: %w", name, err)
		}
		step, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid step value for %s: %w", name, err)
		}

		// Determine type (int or float)
		paramType := "float"
		if step == float64(int(step)) && start == float64(int(start)) && end == float64(int(end)) {
			paramType = "int"
		}

		// Path is same as name for now
		ranges = append(ranges, optimization.ParameterRange{
			Name:  name,
			Start: start,
			End:   end,
			Step:  step,
			Type:  paramType,
			Path:  name,
		})
	}

	return ranges, nil
}

func runWalkforward(args []string) error {
	fs := flag.NewFlagSet("walkforward", flag.ExitOnError)
	strategyPath := fs.String("strategy", "", "Path to strategy YAML file")
	dataPath := fs.String("data", "", "Path to market data file")
	symbol := fs.String("symbol", "BTCUSDT", "Trading symbol")
	initialCash := fs.Float64("cash", 10000, "Initial cash")
	trainBars := fs.Int("train", 1000, "Number of bars for training period")
	testBars := fs.Int("test", 200, "Number of bars for testing (out-of-sample) period")
	stepBars := fs.Int("step", 0, "Number of bars to step forward (0 = step = test)")
	outputJSON := fs.String("output", "", "Output JSON file path")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if *strategyPath == "" {
		return fmt.Errorf("--strategy flag required")
	}
	if *dataPath == "" {
		return fmt.Errorf("--data flag required")
	}

	// Parse strategy to get timeframe
	p := parser.NewParser()
	strategyAST, err := p.ParseFile(*strategyPath)
	if err != nil {
		return fmt.Errorf("parse strategy: %w", err)
	}

	// Create walk forward config
	config := walkforward.WindowConfig{
		TrainBars: *trainBars,
		TestBars:  *testBars,
		StepBars:  *stepBars,
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid walk forward config: %w", err)
	}

	// Load data to determine total bars
	csvConfig := csv.DefaultCSVConfig(market.Symbol(*symbol), market.Timeframe(strategyAST.Data.Timeframe))
	csvData, err := csv.NewCSVFeed(*dataPath, csvConfig)
	if err != nil {
		return fmt.Errorf("load data: %w", err)
	}

	// Get total bars
	totalBars := csvData.Length()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("WALK FORWARD ANALYSIS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Strategy:       %s\n", *strategyPath)
	fmt.Printf("Symbol:         %s\n", *symbol)
	fmt.Printf("Timeframe:      %s\n", strategyAST.Data.Timeframe)
	fmt.Printf("Total Bars:     %d\n", totalBars)
	fmt.Printf("Train Bars:     %d\n", *trainBars)
	fmt.Printf("Test Bars:      %d\n", *testBars)
	fmt.Printf("Step Bars:      %d\n", config.StepBars)
	fmt.Printf("Initial Cash:   $%.2f\n", *initialCash)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create WFA instance
	wfa, err := walkforward.New(config)
	if err != nil {
		return fmt.Errorf("create walk forward analysis: %w", err)
	}

	// Generate windows
	if err := wfa.GenerateWindows(totalBars); err != nil {
		return fmt.Errorf("generate windows: %w", err)
	}

	fmt.Printf("Windows:        %d\n", wfa.WindowCount())
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// For now, we'll just show the configuration
	if wfa.WindowCount() > 0 {
		fmt.Println("Window Configuration:")
		for i := 0; i < min(5, wfa.WindowCount()); i++ {
			window := wfa.GetWindow(i)
			if window != nil {
				fmt.Printf("  Window %d: Train [%d-%d], Test [%d-%d]\n", 
					window.WindowID, 
					window.TrainStart, 
					window.TrainEnd,
					window.TestStart,
					window.TestEnd)
			}
		}
		if wfa.WindowCount() > 5 {
			fmt.Printf("  ... and %d more windows\n", wfa.WindowCount()-5)
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🚧 Walk Forward Analysis engine is ready.")
	fmt.Println("Backtest execution for each window will be implemented next.")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Export results if requested
	if *outputJSON != "" {
		// For now, save basic configuration
		output := map[string]interface{}{
			"strategy":       *strategyPath,
			"symbol":         *symbol,
			"timeframe":      strategyAST.Data.Timeframe,
			"total_bars":     totalBars,
			"train_bars":     *trainBars,
			"test_bars":      *testBars,
			"step_bars":      config.StepBars,
			"window_count":   wfa.WindowCount(),
			"windows":        wfa.Windows,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}

		if err := os.WriteFile(*outputJSON, jsonBytes, 0644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}

		fmt.Printf("Configuration saved to: %s\n", *outputJSON)
	}

	return nil
}

func runMonteCarlo(args []string) error {
	fs := flag.NewFlagSet("montecarlo", flag.ExitOnError)
	resultPath := fs.String("result", "", "Path to backtest result JSON file")
	simulations := fs.Int("simulations", 1000, "Number of Monte Carlo simulations")
	seed := fs.Int64("seed", 42, "Random seed for reproducibility")
	outputJSON := fs.String("output", "", "Output JSON file path")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if *resultPath == "" {
		return fmt.Errorf("--result flag required")
	}

	// Load backtest result
	resultBytes, err := os.ReadFile(*resultPath)
	if err != nil {
		return fmt.Errorf("read result file: %w", err)
	}

	var result backtest.BacktestResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return fmt.Errorf("unmarshal result: %w", err)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("MONTE CARLO SIMULATION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Strategy:     %s\n", result.StrategyName)
	fmt.Printf("Symbol:       %s\n", result.Config.Symbol)
	fmt.Printf("Timeframe:    %s\n", result.Config.Timeframe)
	fmt.Printf("Trades:       %d\n", result.TotalTrades)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Simulations:  %d\n", *simulations)
	fmt.Printf("Seed:         %d\n", *seed)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Convert trades to montecarlo format
	mcTrades := make([]montecarlo.Trade, len(result.TradeHistory))
	for i, t := range result.TradeHistory {
		mcTrades[i] = montecarlo.Trade{
			ID:         int64(i),
			EntryTime:  t.EntryTime,
			ExitTime:   t.ExitTime,
			EntryPrice: t.EntryPrice,
			ExitPrice:  t.ExitPrice,
			Quantity:   t.Quantity,
			GrossPnL:   t.GrossPnL,
			Fees:       t.Fees,
			NetPnL:     t.NetPnL,
			Return:     t.Return,
			MAE:        t.MAE,
			MFE:        t.MFE,
			Duration:   t.ExitTime.Sub(t.EntryTime),
		}
	}

	// Create Monte Carlo runner
	runner := montecarlo.NewRunner(montecarlo.MCConfig{
		Simulations: *simulations,
		Seed:        *seed,
		Type:        montecarlo.TradeReshuffle,
	}, mcTrades, result.Config.InitialCash)

	// Run Monte Carlo simulation
	fmt.Println("\nRunning Monte Carlo simulation...")
	start := time.Now()

	mcResult, err := runner.Run()
	if err != nil {
		return fmt.Errorf("Monte Carlo failed: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("Completed in %v\n", elapsed)

	// Display Monte Carlo results
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("MONTE CARLO ANALYSIS RESULTS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get Monte Carlo results
	stats := mcResult.Statistics

	fmt.Printf("Mean Return:     %+.2f%%\n", stats.MeanReturn*100)
	fmt.Printf("Std Dev:         %.2f%%\n", stats.StdDevReturn*100)
	fmt.Printf("Sharpe Ratio:    %.2f\n", stats.MeanSharpe)
	fmt.Println()

	fmt.Printf("5th Percentile:  %+.2f%%\n", stats.P05Return*100)
	fmt.Printf("50th Percentile: %+.2f%%\n", stats.MedianReturn*100)
	fmt.Printf("95th Percentile: %+.2f%%\n", stats.P95Return*100)
	fmt.Println()

	fmt.Printf("Probability of Loss:   %.2f%%\n", stats.NegativeReturnRatio*100)
	fmt.Printf("Probability of Gain:   %.2f%%\n", (1-stats.NegativeReturnRatio)*100)
	fmt.Printf("Probability of Ruin:   %.2f%%\n", stats.ProbabilityOfRuin*100)
	fmt.Println()

	fmt.Printf("Mean Max Drawdown:     %.2f%%\n", stats.MeanMaxDrawdown*100)
	fmt.Printf("95th Pctl Drawdown:    %.2f%%\n", stats.P95MaxDrawdown*100)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Export to JSON if requested
	if *outputJSON != "" {
		output := map[string]interface{}{
			"strategy":             result.StrategyName,
			"symbol":               result.Config.Symbol,
			"timeframe":            result.Config.Timeframe,
			"trades":               result.TotalTrades,
			"simulations":          *simulations,
			"seed":                 *seed,
			"mean_return":          stats.MeanReturn,
			"std_dev":              stats.StdDevReturn,
			"mean_sharpe":          stats.MeanSharpe,
			"percentile_5":         stats.P05Return,
			"percentile_50":        stats.MedianReturn,
			"percentile_95":        stats.P95Return,
			"probability_of_loss":  stats.NegativeReturnRatio,
			"probability_of_gain":  1 - stats.NegativeReturnRatio,
			"probability_of_ruin":  stats.ProbabilityOfRuin,
			"mean_max_drawdown":    stats.MeanMaxDrawdown,
			"percentile_95_dd":     stats.P95MaxDrawdown,
			"confidence_interval": []float64{
				stats.P05Return,
				stats.P95Return,
			},
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}

		if err := os.WriteFile(*outputJSON, jsonBytes, 0644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}

		fmt.Printf("\nResults saved to: %s\n", *outputJSON)
	}

	return nil
}

func runReport(args []string) error {
	fmt.Println("Report generation not yet implemented")
	return nil
}

func runPaper(args []string) error {
	fs := flag.NewFlagSet("paper", flag.ExitOnError)
	strategyPath := fs.String("strategy", "", "Path to strategy YAML file")
	symbol := fs.String("symbol", "BTCUSDT", "Symbol to trade")
	initialPrice := fs.Float64("price", 50000.0, "Initial price")
	initialBalance := fs.Float64("balance", 10000.0, "Initial balance")
	duration := fs.Int("duration", 60, "Duration in seconds")
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	// Load and validate strategy
	if *strategyPath == "" {
		return fmt.Errorf("strategy file required (--strategy)")
	}

	strategyData, err := os.ReadFile(*strategyPath)
	if err != nil {
		return fmt.Errorf("read strategy file: %w", err)
	}

	p := parser.NewParser()
	strategy, err := p.Parse(strategyData)
	if err != nil {
		return fmt.Errorf("parse strategy: %w", err)
	}

	fmt.Printf("Starting paper trading...\n")
	fmt.Printf("Strategy: %s\n", strategy.Name)
	fmt.Printf("Symbol: %s\n", *symbol)
	fmt.Printf("Initial Price: %.2f\n", *initialPrice)
	fmt.Printf("Initial Balance: %.2f\n", *initialBalance)
	fmt.Printf("Duration: %d seconds\n", *duration)
	fmt.Printf("\nPress Ctrl+C to stop\n\n")

	// Create paper broker
	executor := execution.NewSimpleExecutor(execution.Config{})
	port := portfolio.NewPortfolio(*initialBalance)
	broker := broker.NewPaperBroker(executor, port, broker.DefaultLatencyConfig())
	defer broker.Close()

	// Set initial price
	broker.UpdatePrice(*symbol, *initialPrice)

	// Simple price simulation: static price for MVP
	// TODO: Add price feed with random walk or live data
	
	startTime := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime).Seconds()
			if elapsed >= float64(*duration) {
				fmt.Println("\nPaper trading session complete")
				return printPaperSummary(broker)
			}

			// Print status
			ctx := context.Background()
			positions, _ := broker.GetPositions(ctx)
			balance, _ := broker.GetBalance(ctx)

			fmt.Printf("[%.0fs] Balance: %.2f | Equity: %.2f | Positions: %d\n",
				elapsed, balance.Cash, balance.Equity, len(positions))

			for _, pos := range positions {
				pnl := pos.UnrealizedPnL()
				fmt.Printf("  %s: %.4f @ %.2f (PnL: %.2f)\n",
					pos.Symbol, pos.Quantity, pos.EntryPrice, pnl)
			}
		}
	}
}

func printPaperSummary(broker *broker.PaperBroker) error {
	ctx := context.Background()
	
	positions, err := broker.GetPositions(ctx)
	if err != nil {
		return fmt.Errorf("get positions: %w", err)
	}

	balance, err := broker.GetBalance(ctx)
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("PAPER TRADING SUMMARY")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("\nFinal Balance: %.2f\n", balance.Cash)
	fmt.Printf("Final Equity:  %.2f\n", balance.Equity)
	fmt.Printf("Positions:     %d\n", len(positions))

	if len(positions) > 0 {
		fmt.Println("\nOpen Positions:")
		for _, pos := range positions {
			pnl := pos.UnrealizedPnL()
			fmt.Printf("  %s: %.4f @ %.2f (Unrealized PnL: %.2f)\n",
				pos.Symbol, pos.Quantity, pos.EntryPrice, pnl)
		}
	}

	fmt.Println()
	return nil
}
