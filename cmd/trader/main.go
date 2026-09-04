package main

import (
	"fmt"
	"os"
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
	fmt.Println("Validation not yet implemented")
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
