package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/hserkanyilmaz/costdiff/internal/diff"
)

// RenderCSV outputs the diff result as CSV to stdout
func RenderCSV(result *diff.Result) error {
	return RenderCSVTo(os.Stdout, result)
}

// RenderCSVTo outputs the diff result as CSV to the specified writer
func RenderCSVTo(w io.Writer, result *diff.Result) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"name",
		"from_period",
		"to_period",
		"from_cost",
		"to_cost",
		"diff",
		"diff_percent",
		"is_new",
		"is_removed",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, item := range result.Items {
		row := []string{
			item.Name,
			result.FromPeriod.Label(),
			result.ToPeriod.Label(),
			fmt.Sprintf("%.2f", item.FromCost),
			fmt.Sprintf("%.2f", item.ToCost),
			fmt.Sprintf("%.2f", item.Diff),
			fmt.Sprintf("%.2f", item.DiffPct),
			fmt.Sprintf("%t", item.IsNew),
			fmt.Sprintf("%t", item.IsRemoved),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// RenderTopCSV outputs the top result as CSV to stdout
func RenderTopCSV(result *diff.TopResult) error {
	return RenderTopCSVTo(os.Stdout, result)
}

// RenderTopCSVTo outputs the top result as CSV to the specified writer
func RenderTopCSVTo(w io.Writer, result *diff.TopResult) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"rank", "name", "period", "cost", "percent"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for i, item := range result.Items {
		row := []string{
			fmt.Sprintf("%d", i+1),
			item.Name,
			result.Period.Label(),
			fmt.Sprintf("%.2f", item.Cost),
			fmt.Sprintf("%.2f", item.Percent),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// RenderWatchCSV outputs the watch result as CSV to stdout
func RenderWatchCSV(result *diff.WatchResult) error {
	return RenderWatchCSVTo(os.Stdout, result)
}

// RenderWatchCSVTo outputs the watch result as CSV to the specified writer
func RenderWatchCSVTo(w io.Writer, result *diff.WatchResult) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"date", "day_of_week", "cost", "change", "change_percent"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, day := range result.Days {
		row := []string{
			day.Date.Format("2006-01-02"),
			day.Date.Format("Monday"),
			fmt.Sprintf("%.2f", day.Cost),
			fmt.Sprintf("%.2f", day.Change),
			fmt.Sprintf("%.2f", day.ChangePercent),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

