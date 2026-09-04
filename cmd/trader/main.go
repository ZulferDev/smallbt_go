package main

import (
	"flag"
	"fmt"
	"os"

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
	fmt.Println("Backtest not yet implemented")
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
