package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/hserkanyilmaz/costdiff/internal/aws"
	"github.com/hserkanyilmaz/costdiff/internal/diff"
	"github.com/hserkanyilmaz/costdiff/internal/output"
)

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Show top cost drivers for current period",
	Long: `Show the top cost drivers for the current month (or specified period).

Examples:
  costdiff top                # Top 10 services this month
  costdiff top -n 20          # Top 20 services
  costdiff top -g region      # Top costs by region
  costdiff top --from 2024-10 # Top costs for October 2024`,
	RunE: runTop,
}

func init() {
	rootCmd.AddCommand(topCmd)
}

func runTop(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	// Parse time period
	period, err := parseTopPeriod(fromPeriod)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}

	debugf("Period: %s to %s", period.Start, period.End)

	// Validate grouping
	groupType, err := parseGroupBy(groupBy, tagKey)
	if err != nil {
		return err
	}

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

	// Fetch cost data with spinner
	costs, err := withSpinner("Fetching cost data...", func() (map[string]float64, error) {
		return client.GetCosts(ctx, period.Start, period.End, groupType, metric)
	})
	if err != nil {
		return handleAWSError(err)
	}

	// Build result
	result := buildTopResult(costs, period)

	// Apply threshold filter
	if threshold > 0 {
		result = filterTopByThreshold(result, threshold)
	}

	// Limit results
	if len(result.Items) > topN {
		result.Items = result.Items[:topN]
	}

	// Output
	return outputTopResult(result, outputFmt)
}

func parseTopPeriod(from string) (diff.Period, error) {
	now := time.Now()

	if from == "" {
		// Default: current month
		nextMonth := now.AddDate(0, 1, 0)
		return diff.Period{
			Start: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC),
		}, nil
	}

	return parseDate(from)
}

func buildTopResult(costs map[string]float64, period diff.Period) *diff.TopResult {
	var items []diff.TopItem
	var total float64

	for name, cost := range costs {
		items = append(items, diff.TopItem{
			Name: name,
			Cost: cost,
		})
		total += cost
	}

	// Sort by cost descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].Cost > items[j].Cost
	})

	// Calculate percentages
	for i := range items {
		if total > 0 {
			items[i].Percent = (items[i].Cost / total) * 100
		}
	}

	return &diff.TopResult{
		Period: period,
		Total:  total,
		Items:  items,
	}
}

func filterTopByThreshold(result *diff.TopResult, threshold float64) *diff.TopResult {
	filtered := &diff.TopResult{
		Period: result.Period,
		Total:  result.Total,
		Items:  make([]diff.TopItem, 0),
	}

	for _, item := range result.Items {
		if item.Cost >= threshold {
			filtered.Items = append(filtered.Items, item)
		}
	}

	return filtered
}

func outputTopResult(result *diff.TopResult, format string) error {
	switch format {
	case "table":
		return output.RenderTopTable(result)
	case "json":
		return output.RenderTopJSON(result)
	case "csv":
		return output.RenderTopCSV(result)
	default:
		return fmt.Errorf("invalid output format: %s (must be table|json|csv)", format)
	}
}
