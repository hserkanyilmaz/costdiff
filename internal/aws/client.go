package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
)

// CostExplorerClient wraps the AWS Cost Explorer client
type CostExplorerClient struct {
	client *costexplorer.Client
}

// NewCostExplorerClient creates a new Cost Explorer client with the given profile and region
func NewCostExplorerClient(ctx context.Context, profile, region string) (*CostExplorerClient, error) {
	var opts []func(*config.LoadOptions) error

	// Use profile if specified
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	// Use region if specified, otherwise default to us-east-1 for Cost Explorer
	// Cost Explorer is a global service but requires a region
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	} else {
		opts = append(opts, config.WithRegion("us-east-1"))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := costexplorer.NewFromConfig(cfg)

	return &CostExplorerClient{
		client: client,
	}, nil
}

// GroupType defines how to group cost data
type GroupType struct {
	Type string // DIMENSION or TAG
	Key  string // SERVICE, REGION, LINKED_ACCOUNT, or tag key
}

// Predefined group types
var (
	GroupByService = GroupType{Type: "DIMENSION", Key: "SERVICE"}
	GroupByRegion  = GroupType{Type: "DIMENSION", Key: "REGION"}
	GroupByAccount = GroupType{Type: "DIMENSION", Key: "LINKED_ACCOUNT"}
)

// GetClient returns the underlying Cost Explorer client for advanced use cases
func (c *CostExplorerClient) GetClient() *costexplorer.Client {
	return c.client
}

// FormatCurrency formats a float as a dollar amount
func FormatCurrency(amount float64) string {
	if amount < 0 {
		return fmt.Sprintf("-$%.2f", -amount)
	}
	return fmt.Sprintf("$%.2f", amount)
}

// FormatPercent formats a float as a percentage
func FormatPercent(pct float64) string {
	if pct >= 0 {
		return fmt.Sprintf("+%.1f%%", pct)
	}
	return fmt.Sprintf("%.1f%%", pct)
}

// FormatChange formats a cost change with sign
func FormatChange(change float64) string {
	if change >= 0 {
		return fmt.Sprintf("+$%.2f", change)
	}
	return fmt.Sprintf("-$%.2f", -change)
}

// Ptr returns a pointer to the given string
func Ptr(s string) *string {
	return aws.String(s)
}

