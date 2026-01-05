package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
)

// CostFetcher defines the interface for fetching AWS cost data.
// This interface allows for easy mocking in tests.
type CostFetcher interface {
	// GetCosts fetches cost data for a given period grouped by the specified type.
	GetCosts(ctx context.Context, start, end time.Time, groupBy GroupType, metric string) (map[string]float64, error)

	// GetDailyCosts fetches daily cost data for a given period.
	GetDailyCosts(ctx context.Context, start, end time.Time, metric string) ([]DailyCost, error)

	// GetTotalCost fetches total cost for a period without grouping.
	GetTotalCost(ctx context.Context, start, end time.Time, metric string) (float64, error)

	// SetLogger sets the logger for the client.
	SetLogger(logger Logger)
}

// Ensure CostExplorerClient implements CostFetcher
var _ CostFetcher = (*CostExplorerClient)(nil)

// Logger interface for debug/warning logging
type Logger interface {
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

// noopLogger is a logger that does nothing
type noopLogger struct{}

func (noopLogger) Debugf(format string, args ...interface{}) {}
func (noopLogger) Warnf(format string, args ...interface{})  {}

// CostExplorerClient wraps the AWS Cost Explorer client
type CostExplorerClient struct {
	client *costexplorer.Client
	logger Logger
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
		logger: noopLogger{},
	}, nil
}

// SetLogger sets the logger for the client
func (c *CostExplorerClient) SetLogger(logger Logger) {
	if logger != nil {
		c.logger = logger
	}
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
