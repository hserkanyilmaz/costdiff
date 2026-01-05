package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	fromPeriod string
	toPeriod   string
	groupBy    string
	tagKey     string
	topN       int
	outputFmt  string
	awsProfile string
	awsRegion  string
	threshold  float64
	costMetric string
	quiet      bool
	verbose    bool
)

// Valid cost metrics
var validMetrics = map[string]string{
	"net-amortized":  "NetAmortizedCost",
	"amortized":      "AmortizedCost",
	"unblended":      "UnblendedCost",
	"blended":        "BlendedCost",
	"net-unblended":  "NetUnblendedCost",
	"normalized":     "NormalizedUsageAmount",
	"usage-quantity": "UsageQuantity",
}

var rootCmd = &cobra.Command{
	Use:   "costdiff",
	Short: "Compare AWS costs between time periods",
	Long: `costdiff is a CLI tool that compares AWS costs between two time periods.

It helps you identify cost changes, top cost drivers, and daily trends
in your AWS spending using the Cost Explorer API.

Examples:
  costdiff                              # Compare last month vs current month
  costdiff --from 2024-10 --to 2024-12  # Compare specific months
  costdiff -g tag --tag team            # Group by tag
  costdiff top                          # Show top cost drivers
  costdiff watch                        # Show daily cost trend`,
	RunE: runDiff,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Time period flags
	rootCmd.PersistentFlags().StringVarP(&fromPeriod, "from", "f", "", "Start period (YYYY-MM or YYYY-MM-DD)")
	rootCmd.PersistentFlags().StringVarP(&toPeriod, "to", "t", "", "End period (YYYY-MM or YYYY-MM-DD)")

	// Grouping flags
	rootCmd.PersistentFlags().StringVarP(&groupBy, "group", "g", "service", "Group by: service|tag|region|account")
	rootCmd.PersistentFlags().StringVar(&tagKey, "tag", "", "Tag key when grouping by tag")

	// Output flags
	rootCmd.PersistentFlags().IntVarP(&topN, "top", "n", 10, "Number of results to show")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "format", "o", "table", "Output format: table|json|csv")

	// AWS flags
	rootCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", "", "AWS profile name")
	rootCmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "AWS region")

	// Filter flags
	rootCmd.PersistentFlags().Float64Var(&threshold, "threshold", 0, "Only show changes above $X")

	// Cost metric flag
	rootCmd.PersistentFlags().StringVarP(&costMetric, "metric", "m", "net-amortized", "Cost metric: net-amortized|amortized|unblended|blended|net-unblended")

	// Verbosity flags
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug output")
}

func debugf(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

func infof(format string, args ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// getAWSMetric converts the user-friendly metric name to AWS API metric name
func getAWSMetric() (string, error) {
	if metric, ok := validMetrics[costMetric]; ok {
		return metric, nil
	}
	return "", fmt.Errorf("invalid metric: %s (valid options: net-amortized, amortized, unblended, blended, net-unblended)", costMetric)
}
