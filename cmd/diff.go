package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	"github.com/hsy/costdiff/internal/aws"
	"github.com/hsy/costdiff/internal/diff"
	"github.com/hsy/costdiff/internal/output"
)

func runDiff(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse time periods
	from, to, err := parsePeriods(fromPeriod, toPeriod)
	if err != nil {
		return fmt.Errorf("invalid date range: %w", err)
	}

	debugf("From period: %s to %s", from.Start, from.End)
	debugf("To period: %s to %s", to.Start, to.End)

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

	// Show spinner for API calls
	var s *spinner.Spinner
	if !quiet {
		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Fetching cost data..."
		s.Start()
	}

	// Fetch cost data for both periods
	fromCosts, err := client.GetCosts(ctx, from.Start, from.End, groupType, metric)
	if err != nil {
		if s != nil {
			s.Stop()
		}
		return handleAWSError(err)
	}

	toCosts, err := client.GetCosts(ctx, to.Start, to.End, groupType, metric)
	if err != nil {
		if s != nil {
			s.Stop()
		}
		return handleAWSError(err)
	}

	if s != nil {
		s.Stop()
	}

	// Calculate diff
	result := diff.Compare(fromCosts, toCosts, from, to)

	// Apply filters
	if threshold > 0 {
		result = filterByThreshold(result, threshold)
	}

	// Limit results
	if len(result.Items) > topN {
		result.Items = result.Items[:topN]
	}

	// Output
	return outputResult(result, outputFmt)
}

func parsePeriods(from, to string) (diff.Period, diff.Period, error) {
	now := time.Now()

	var fromPeriod, toPeriod diff.Period

	if from == "" {
		// Default: last month
		lastMonth := now.AddDate(0, -1, 0)
		fromPeriod = diff.Period{
			Start: time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		}
	} else {
		var err error
		fromPeriod, err = parseDate(from)
		if err != nil {
			return diff.Period{}, diff.Period{}, fmt.Errorf("invalid --from date: %w", err)
		}
	}

	if to == "" {
		// Default: current month
		nextMonth := now.AddDate(0, 1, 0)
		toPeriod = diff.Period{
			Start: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC),
		}
	} else {
		var err error
		toPeriod, err = parseDate(to)
		if err != nil {
			return diff.Period{}, diff.Period{}, fmt.Errorf("invalid --to date: %w", err)
		}
	}

	return fromPeriod, toPeriod, nil
}

func parseDate(s string) (diff.Period, error) {
	// Try YYYY-MM-DD format
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return diff.Period{
			Start: t,
			End:   t.AddDate(0, 0, 1),
		}, nil
	}

	// Try YYYY-MM format (full month)
	if t, err := time.Parse("2006-01", s); err == nil {
		return diff.Period{
			Start: t,
			End:   t.AddDate(0, 1, 0),
		}, nil
	}

	return diff.Period{}, fmt.Errorf("date must be YYYY-MM or YYYY-MM-DD format")
}

func parseGroupBy(group, tag string) (aws.GroupType, error) {
	switch group {
	case "service":
		return aws.GroupByService, nil
	case "region":
		return aws.GroupByRegion, nil
	case "account":
		return aws.GroupByAccount, nil
	case "tag":
		if tag == "" {
			return aws.GroupType{}, fmt.Errorf("--tag is required when grouping by tag")
		}
		return aws.GroupType{Type: "TAG", Key: tag}, nil
	default:
		return aws.GroupType{}, fmt.Errorf("invalid group: %s (must be service|tag|region|account)", group)
	}
}

func filterByThreshold(result *diff.Result, threshold float64) *diff.Result {
	filtered := &diff.Result{
		FromPeriod: result.FromPeriod,
		ToPeriod:   result.ToPeriod,
		FromTotal:  result.FromTotal,
		ToTotal:    result.ToTotal,
		Items:      make([]diff.Item, 0),
	}

	for _, item := range result.Items {
		if abs(item.Diff) >= threshold {
			filtered.Items = append(filtered.Items, item)
		}
	}

	return filtered
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func outputResult(result *diff.Result, format string) error {
	switch format {
	case "table":
		return output.RenderTable(result)
	case "json":
		return output.RenderJSON(result)
	case "csv":
		return output.RenderCSV(result)
	default:
		return fmt.Errorf("invalid output format: %s (must be table|json|csv)", format)
	}
}

func handleAWSError(err error) error {
	errStr := err.Error()

	// Check for common AWS errors
	if contains(errStr, "NoCredentialProviders") || contains(errStr, "no EC2 IMDS role found") {
		return fmt.Errorf("AWS credentials not found.\n\nPlease configure credentials using one of:\n  - AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables\n  - AWS credentials file (~/.aws/credentials)\n  - IAM role (when running on EC2/ECS/Lambda)\n  - Use --profile flag to specify a named profile")
	}

	if contains(errStr, "AccessDeniedException") {
		return fmt.Errorf("access denied. Ensure your IAM user/role has the following permissions:\n  - ce:GetCostAndUsage\n  - ce:GetCostForecast")
	}

	if contains(errStr, "OptInRequired") || contains(errStr, "Cost Explorer has not been enabled") {
		return fmt.Errorf("AWS Cost Explorer is not enabled for this account.\n\nTo enable it:\n  1. Go to AWS Console > Billing > Cost Explorer\n  2. Click 'Enable Cost Explorer'\n  3. Wait up to 24 hours for data to be available")
	}

	if contains(errStr, "InvalidParameterValue") {
		return fmt.Errorf("invalid parameter: %w\n\nCheck your date range and grouping options", err)
	}

	return fmt.Errorf("AWS API error: %w", err)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
