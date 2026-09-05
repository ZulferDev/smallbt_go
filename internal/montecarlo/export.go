package montecarlo

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ExportToJSON exports Monte Carlo results to JSON file
func (r *MCResult) ExportToJSON(filepath string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// ExportToCSV exports key statistics to CSV file
func (r *MCResult) ExportToCSV(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write configuration
	if err := writer.Write([]string{"Configuration", ""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Simulations", fmt.Sprintf("%d", r.Config.Simulations)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Seed", fmt.Sprintf("%d", r.Config.Seed)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Type", r.Config.Type.String()}); err != nil {
		return err
	}
	if err := writer.Write([]string{""}); err != nil {
		return err
	}

	// Write statistics header
	if err := writer.Write([]string{"Statistics", ""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Metric", "Value"}); err != nil {
		return err
	}

	// Write return statistics
	if err := writer.Write([]string{"Mean Return", fmt.Sprintf("%.4f", r.Statistics.MeanReturn)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Std Dev Return", fmt.Sprintf("%.4f", r.Statistics.StdDevReturn)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Min Return", fmt.Sprintf("%.4f", r.Statistics.MinReturn)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Max Return", fmt.Sprintf("%.4f", r.Statistics.MaxReturn)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Median Return", fmt.Sprintf("%.4f", r.Statistics.MedianReturn)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"P05 Return", fmt.Sprintf("%.4f", r.Statistics.P05Return)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"P95 Return", fmt.Sprintf("%.4f", r.Statistics.P95Return)}); err != nil {
		return err
	}
	if err := writer.Write([]string{""}); err != nil {
		return err
	}

	// Write drawdown statistics
	if err := writer.Write([]string{"Mean Max Drawdown", fmt.Sprintf("%.4f", r.Statistics.MeanMaxDrawdown)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Std Dev Max Drawdown", fmt.Sprintf("%.4f", r.Statistics.StdDevMaxDrawdown)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Min Max Drawdown", fmt.Sprintf("%.4f", r.Statistics.MinMaxDrawdown)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Max Max Drawdown", fmt.Sprintf("%.4f", r.Statistics.MaxMaxDrawdown)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Median Max Drawdown", fmt.Sprintf("%.4f", r.Statistics.MedianMaxDrawdown)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"P95 Max Drawdown (Worst Case)", fmt.Sprintf("%.4f", r.Statistics.P95MaxDrawdown)}); err != nil {
		return err
	}
	if err := writer.Write([]string{""}); err != nil {
		return err
	}

	// Write win rate statistics
	if err := writer.Write([]string{"Mean Win Rate", fmt.Sprintf("%.4f", r.Statistics.MeanWinRate)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Std Dev Win Rate", fmt.Sprintf("%.4f", r.Statistics.StdDevWinRate)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Min Win Rate", fmt.Sprintf("%.4f", r.Statistics.MinWinRate)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Max Win Rate", fmt.Sprintf("%.4f", r.Statistics.MaxWinRate)}); err != nil {
		return err
	}
	if err := writer.Write([]string{""}); err != nil {
		return err
	}

	// Write Sharpe ratio statistics
	if err := writer.Write([]string{"Mean Sharpe", fmt.Sprintf("%.4f", r.Statistics.MeanSharpe)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Std Dev Sharpe", fmt.Sprintf("%.4f", r.Statistics.StdDevSharpe)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Min Sharpe", fmt.Sprintf("%.4f", r.Statistics.MinSharpe)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Max Sharpe", fmt.Sprintf("%.4f", r.Statistics.MaxSharpe)}); err != nil {
		return err
	}
	if err := writer.Write([]string{""}); err != nil {
		return err
	}

	// Write probability statistics
	if err := writer.Write([]string{"Probability of Ruin", fmt.Sprintf("%.4f", r.Statistics.ProbabilityOfRuin)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Negative Return Count", fmt.Sprintf("%d", r.Statistics.NegativeReturnCount)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Negative Return Ratio", fmt.Sprintf("%.4f", r.Statistics.NegativeReturnRatio)}); err != nil {
		return err
	}
	if err := writer.Write([]string{""}); err != nil {
		return err
	}

	// Write confidence intervals header
	if err := writer.Write([]string{"Confidence Intervals", ""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Percentile", "Total Return", "Max Drawdown", "Win Rate", "Sharpe Ratio"}); err != nil {
		return err
	}

	// Write confidence intervals
	for _, ci := range r.ConfidenceIntervals {
		if err := writer.Write([]string{
			fmt.Sprintf("%.0f%%", ci.Percentile*100),
			fmt.Sprintf("%.4f", ci.TotalReturn),
			fmt.Sprintf("%.4f", ci.MaxDrawdown),
			fmt.Sprintf("%.4f", ci.WinRate),
			fmt.Sprintf("%.4f", ci.SharpeRatio),
		}); err != nil {
			return err
		}
	}

	return nil
}

// ExportToText exports results as human-readable text
func (r *MCResult) ExportToText() string {
	var sb strings.Builder

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("MONTE CARLO SIMULATION RESULTS\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	sb.WriteString(fmt.Sprintf("Config:\n"))
	sb.WriteString(fmt.Sprintf("  Simulations:     %d\n", r.Config.Simulations))
	sb.WriteString(fmt.Sprintf("  Seed:            %d\n", r.Config.Seed))
	sb.WriteString(fmt.Sprintf("  Type:            %s\n\n", r.Config.Type))

	sb.WriteString("Return Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Mean Return:     %.4f\n", r.Statistics.MeanReturn))
	sb.WriteString(fmt.Sprintf("  Std Dev:         %.4f\n", r.Statistics.StdDevReturn))
	sb.WriteString(fmt.Sprintf("  Min/Max:         [%.4f, %.4f]\n", r.Statistics.MinReturn, r.Statistics.MaxReturn))
	sb.WriteString(fmt.Sprintf("  Median:          %.4f\n", r.Statistics.MedianReturn))
	sb.WriteString(fmt.Sprintf("  5th / 95th:      [%.4f, %.4f]\n\n", r.Statistics.P05Return, r.Statistics.P95Return))

	sb.WriteString("Drawdown Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Mean Drawdown:   %.4f\n", r.Statistics.MeanMaxDrawdown))
	sb.WriteString(fmt.Sprintf("  Std Dev:         %.4f\n", r.Statistics.StdDevMaxDrawdown))
	sb.WriteString(fmt.Sprintf("  Min/Max:         [%.4f, %.4f]\n", r.Statistics.MinMaxDrawdown, r.Statistics.MaxMaxDrawdown))
	sb.WriteString(fmt.Sprintf("  Median:          %.4f\n", r.Statistics.MedianMaxDrawdown))
	sb.WriteString(fmt.Sprintf("  95th (Worst):    %.4f\n\n", r.Statistics.P95MaxDrawdown))

	sb.WriteString("Win Rate Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Mean Win Rate:   %.4f\n", r.Statistics.MeanWinRate))
	sb.WriteString(fmt.Sprintf("  Std Dev:         %.4f\n", r.Statistics.StdDevWinRate))
	sb.WriteString(fmt.Sprintf("  Min/Max:         [%.4f, %.4f]\n\n", r.Statistics.MinWinRate, r.Statistics.MaxWinRate))

	sb.WriteString("Sharpe Ratio Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Mean Sharpe:     %.4f\n", r.Statistics.MeanSharpe))
	sb.WriteString(fmt.Sprintf("  Std Dev:         %.4f\n", r.Statistics.StdDevSharpe))
	sb.WriteString(fmt.Sprintf("  Min/Max:         [%.4f, %.4f]\n\n", r.Statistics.MinSharpe, r.Statistics.MaxSharpe))

	sb.WriteString("Probability Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Probability of Ruin:     %.4f\n", r.Statistics.ProbabilityOfRuin))
	sb.WriteString(fmt.Sprintf("  Negative Return Count:   %d\n", r.Statistics.NegativeReturnCount))
	sb.WriteString(fmt.Sprintf("  Negative Return Ratio:   %.4f\n\n", r.Statistics.NegativeReturnRatio))

	sb.WriteString("Confidence Intervals:\n")
	sb.WriteString(fmt.Sprintf("  Percentile  Return     Drawdown   Win Rate   Sharpe\n"))
	for _, ci := range r.ConfidenceIntervals {
		sb.WriteString(fmt.Sprintf("  %6.0f%%   %+8.4f   %+8.4f   %+8.4f   %+8.4f\n",
			ci.Percentile*100, ci.TotalReturn, ci.MaxDrawdown, ci.WinRate, ci.SharpeRatio))
	}

	return sb.String()
}

// String returns a summary of the Monte Carlo results
func (r *MCResult) String() string {
	return fmt.Sprintf("Monte Carlo: %d simulations, Mean Return: %.4f, Mean Drawdown: %.4f, Prob Ruin: %.4f",
		r.Config.Simulations, r.Statistics.MeanReturn, r.Statistics.MeanMaxDrawdown, r.Statistics.ProbabilityOfRuin)
}
