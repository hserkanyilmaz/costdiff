package aws

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// DailyCost represents cost for a single day
type DailyCost struct {
	Date time.Time
	Cost float64
}

// GetCosts fetches cost data for a given period grouped by the specified type
func (c *CostExplorerClient) GetCosts(ctx context.Context, start, end time.Time, groupBy GroupType, metric string) (map[string]float64, error) {
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(end.Format("2006-01-02")),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{metric},
		GroupBy:     buildGroupDefinition(groupBy),
	}

	result, err := c.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost data: %w", err)
	}

	costs := make(map[string]float64)

	for _, resultByTime := range result.ResultsByTime {
		for _, group := range resultByTime.Groups {
			name := getGroupName(group.Keys)
			amount := parseAmount(group.Metrics[metric])
			costs[name] += amount
		}
	}

	return costs, nil
}

// GetDailyCosts fetches daily cost data for a given period
func (c *CostExplorerClient) GetDailyCosts(ctx context.Context, start, end time.Time, metric string) ([]DailyCost, error) {
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(end.Format("2006-01-02")),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{metric},
	}

	result, err := c.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily cost data: %w", err)
	}

	var dailyCosts []DailyCost

	for _, resultByTime := range result.ResultsByTime {
		date, err := time.Parse("2006-01-02", *resultByTime.TimePeriod.Start)
		if err != nil {
			continue
		}

		var totalCost float64
		if len(resultByTime.Groups) > 0 {
			for _, group := range resultByTime.Groups {
				totalCost += parseAmount(group.Metrics[metric])
			}
		} else {
			totalCost = parseAmount(resultByTime.Total[metric])
		}

		dailyCosts = append(dailyCosts, DailyCost{
			Date: date,
			Cost: totalCost,
		})
	}

	return dailyCosts, nil
}

// GetCostsByService fetches cost data grouped by service with additional breakdown
func (c *CostExplorerClient) GetCostsByService(ctx context.Context, start, end time.Time, metric string) (map[string]float64, error) {
	return c.GetCosts(ctx, start, end, GroupByService, metric)
}

// GetCostsByRegion fetches cost data grouped by region
func (c *CostExplorerClient) GetCostsByRegion(ctx context.Context, start, end time.Time, metric string) (map[string]float64, error) {
	return c.GetCosts(ctx, start, end, GroupByRegion, metric)
}

// GetCostsByAccount fetches cost data grouped by linked account
func (c *CostExplorerClient) GetCostsByAccount(ctx context.Context, start, end time.Time, metric string) (map[string]float64, error) {
	return c.GetCosts(ctx, start, end, GroupByAccount, metric)
}

// GetCostsByTag fetches cost data grouped by a specific tag
func (c *CostExplorerClient) GetCostsByTag(ctx context.Context, start, end time.Time, tagKey, metric string) (map[string]float64, error) {
	return c.GetCosts(ctx, start, end, GroupType{Type: "TAG", Key: tagKey}, metric)
}

// buildGroupDefinition creates the GroupBy definition for the API
func buildGroupDefinition(groupBy GroupType) []types.GroupDefinition {
	var groupType types.GroupDefinitionType

	switch groupBy.Type {
	case "TAG":
		groupType = types.GroupDefinitionTypeTag
	default:
		groupType = types.GroupDefinitionTypeDimension
	}

	return []types.GroupDefinition{
		{
			Type: groupType,
			Key:  aws.String(groupBy.Key),
		},
	}
}

// getGroupName extracts a readable name from group keys
func getGroupName(keys []string) string {
	if len(keys) == 0 {
		return "Unknown"
	}

	name := keys[0]

	// Clean up common AWS service name prefixes
	if name == "" {
		return "Other"
	}

	return name
}

// parseAmount parses a MetricValue to float64
func parseAmount(metric types.MetricValue) float64 {
	if metric.Amount == nil {
		return 0
	}

	amount, err := strconv.ParseFloat(*metric.Amount, 64)
	if err != nil {
		return 0
	}

	return amount
}

// GetTotalCost fetches total cost for a period without grouping
func (c *CostExplorerClient) GetTotalCost(ctx context.Context, start, end time.Time, metric string) (float64, error) {
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(end.Format("2006-01-02")),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{metric},
	}

	result, err := c.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to get total cost: %w", err)
	}

	var total float64
	for _, resultByTime := range result.ResultsByTime {
		total += parseAmount(resultByTime.Total[metric])
	}

	return total, nil
}
