package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/hserkanyilmaz/costdiff/internal/aws"
	"github.com/hserkanyilmaz/costdiff/internal/diff"
	"github.com/hserkanyilmaz/costdiff/internal/output"
)

var (
	watchDays int
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Show daily cost trend",
	Long: `Show daily cost trend over a period of time.

Examples:
  costdiff watch              # Last 7 days
  costdiff watch --days 30    # Last 30 days
  costdiff watch -g service   # Daily breakdown by service`,
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().IntVar(&watchDays, "days", 7, "Number of days to show")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	// Calculate date range
	now := time.Now()
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startDate := endDate.AddDate(0, 0, -watchDays)

	debugf("Watch period: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Get metric
	metric, err := getAWSMetric()
	if err != nil {
		return err
	}
	debugf("Using metric: %s", metric)

	// Initialize AWS client
	client, err := aws.NewCostExplorerClient(ctx, awsProfile, awsRegion)
	if err != nil {
		return handleAWSError(err)
	}
	client.SetLogger(cliLogger{})

	// Fetch daily cost data with spinner
	dailyCosts, err := withSpinner("Fetching daily cost data...", func() ([]aws.DailyCost, error) {
		return client.GetDailyCosts(ctx, startDate, endDate, metric)
	})
	if err != nil {
		return handleAWSError(err)
	}

	// Build result
	result := buildWatchResult(dailyCosts, startDate, endDate)

	// Output
	return outputWatchResult(result, outputFmt)
}

func buildWatchResult(dailyCosts []aws.DailyCost, start, end time.Time) *diff.WatchResult {
	var total float64
	var items []diff.DayItem

	for _, dc := range dailyCosts {
		total += dc.Cost
		items = append(items, diff.DayItem{
			Date: dc.Date,
			Cost: dc.Cost,
		})
	}

	// Calculate day-over-day changes
	for i := 1; i < len(items); i++ {
		prev := items[i-1].Cost
		curr := items[i].Cost
		items[i].Change = curr - prev
		if prev > 0 {
			items[i].ChangePercent = ((curr - prev) / prev) * 100
		}
	}

	// Calculate average
	var avg float64
	if len(items) > 0 {
		avg = total / float64(len(items))
	}

	return &diff.WatchResult{
		StartDate: start,
		EndDate:   end,
		Total:     total,
		Average:   avg,
		Days:      items,
	}
}

func outputWatchResult(result *diff.WatchResult, format string) error {
	switch format {
	case "table":
		return output.RenderWatchTable(result)
	case "json":
		return output.RenderWatchJSON(result)
	case "csv":
		return output.RenderWatchCSV(result)
	default:
		return fmt.Errorf("invalid output format: %s (must be table|json|csv)", format)
	}
}
