package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/1jehuang/backtest/internal/backtest"
	"github.com/1jehuang/backtest/internal/market"
	"github.com/1jehuang/backtest/internal/strategy/parser"
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
  report        Generate reports from backtest results

FLAGS:
  -h, --help    Show this help message
  -v, --version Show version information

EXAMPLES:
  trader validate strategy.yaml
  trader backtest --strategy strategy.yaml --data data.csv
  trader report --result backtest_result.json`)
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
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
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
	fmt.Printf("Strategy       %s\n", "ema_volume_atr") // TODO: Get from result
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
	fmt.Printf("Net PnL        $%.2f\n", result.Portfolio.Equity - result.Config.InitialCash)
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
	
	return nil
}

func runOptimize(args []string) error {
	fmt.Println("Optimization not yet implemented")
	return nil
}

func runWalkforward(args []string) error {
	fmt.Println("Walk Forward Analysis not yet implemented")
	return nil
}

func runReport(args []string) error {
	fmt.Println("Report generation not yet implemented")
	return nil
}
