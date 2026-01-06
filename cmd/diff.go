package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/aws/smithy-go"
	"github.com/spf13/cobra"

	"github.com/hserkanyilmaz/costdiff/internal/aws"
	"github.com/hserkanyilmaz/costdiff/internal/diff"
	"github.com/hserkanyilmaz/costdiff/internal/output"
)

// Default timeout for AWS API calls
const defaultAPITimeout = 2 * time.Minute

func runDiff(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

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
	client.SetLogger(cliLogger{})

	// Fetch cost data for both periods with spinner
	spin := newProgressSpinner("Fetching cost data...")
	defer spin.Stop()

	fromCosts, err := client.GetCosts(ctx, from.Start, from.End, groupType, metric, serviceFilter)
	if err != nil {
		return handleAWSError(err)
	}

	toCosts, err := client.GetCosts(ctx, to.Start, to.End, groupType, metric, serviceFilter)
	if err != nil {
		return handleAWSError(err)
	}

	spin.Stop()

	// Calculate diff
	result := diff.Compare(fromCosts, toCosts, from, to)

	// Apply sorting
	applySorting(result.Items, sortBy)

	// Apply filters
	if threshold > 0 {
		result = filterByThreshold(result, threshold)
	}
	if minCost > 0 {
		result.Items = diff.FilterByMinCost(result.Items, minCost)
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

	// Validate that from period starts before to period
	if !fromPeriod.Start.Before(toPeriod.Start) {
		return diff.Period{}, diff.Period{}, fmt.Errorf("--from date (%s) must be before --to date (%s)",
			fromPeriod.Start.Format("2006-01-02"), toPeriod.Start.Format("2006-01-02"))
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
	case "usage-type":
		return aws.GroupByUsageType, nil
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
		return aws.GroupType{}, fmt.Errorf("invalid group: %s (must be service|usage-type|tag|region|account)", group)
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
		if math.Abs(item.Diff) >= threshold {
			filtered.Items = append(filtered.Items, item)
		}
	}

	return filtered
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

func applySorting(items []diff.Item, sortBy string) {
	switch sortBy {
	case "diff":
		diff.SortByDiff(items)
	case "diff-pct":
		diff.SortByDiffPercent(items)
	case "cost":
		diff.SortByToCost(items)
	case "name":
		diff.SortByName(items)
	default:
		// Default to sorting by diff
		diff.SortByDiff(items)
	}
}

func handleAWSError(err error) error {
	// Check for context timeout/cancellation
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("AWS API request timed out. Check your network connection and try again")
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("AWS API request was cancelled")
	}

	// Check for AWS API errors using proper type assertion
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "AccessDeniedException":
			return fmt.Errorf("access denied. Ensure your IAM user/role has the following permissions:\n  - ce:GetCostAndUsage\n  - ce:GetCostForecast")
		case "OptInRequired":
			return fmt.Errorf("AWS Cost Explorer is not enabled for this account.\n\nTo enable it:\n  1. Go to AWS Console > Billing > Cost Explorer\n  2. Click 'Enable Cost Explorer'\n  3. Wait up to 24 hours for data to be available")
		case "InvalidParameterValue", "ValidationException":
			return fmt.Errorf("invalid parameter: %s\n\nCheck your date range and grouping options", apiErr.ErrorMessage())
		case "ThrottlingException", "RequestLimitExceeded":
			return fmt.Errorf("AWS API rate limit exceeded. Please wait a moment and try again")
		}
		// Return the API error message for other AWS errors
		return fmt.Errorf("AWS API error (%s): %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
	}

	// Fallback to string matching for credential errors (these don't implement APIError)
	errStr := err.Error()
	if strings.Contains(errStr, "NoCredentialProviders") || strings.Contains(errStr, "no EC2 IMDS role found") {
		return fmt.Errorf("AWS credentials not found.\n\nPlease configure credentials using one of:\n  - AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables\n  - AWS credentials file (~/.aws/credentials)\n  - IAM role (when running on EC2/ECS/Lambda)\n  - Use --profile flag to specify a named profile")
	}

	// Also check for string-based "Cost Explorer has not been enabled" in case it comes from a different error type
	if strings.Contains(errStr, "Cost Explorer has not been enabled") {
		return fmt.Errorf("AWS Cost Explorer is not enabled for this account.\n\nTo enable it:\n  1. Go to AWS Console > Billing > Cost Explorer\n  2. Click 'Enable Cost Explorer'\n  3. Wait up to 24 hours for data to be available")
	}

	return fmt.Errorf("AWS API error: %w", err)
}
