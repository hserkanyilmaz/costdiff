package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/hserkanyilmaz/costdiff/internal/diff"
)

// RenderTable outputs the diff result as a formatted table to stdout
func RenderTable(result *diff.Result) error {
	return RenderTableTo(os.Stdout, result)
}

// RenderTableTo outputs the diff result as a formatted table to the specified writer
func RenderTableTo(w io.Writer, result *diff.Result) error {
	// Print header
	fmt.Fprintf(w, "\n%s\n\n", Header(fmt.Sprintf("AWS Cost Diff: %s → %s",
		result.FromPeriod.Label(),
		result.ToPeriod.Label())))

	// Print total
	totalChange := FormatDiffFull(result.TotalDiff, result.TotalPct, false, false)
	fmt.Fprintf(w, "Total: %s → %s (%s)\n\n",
		FormatCurrency(result.FromTotal),
		FormatCurrency(result.ToTotal),
		totalChange)

	if len(result.Items) == 0 {
		fmt.Fprintln(w, Muted("No cost data found for the specified period."))
		return nil
	}

	// Create table
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{
		"Service",
		result.FromPeriod.Label(),
		result.ToPeriod.Label(),
		"Change",
	})

	// Configure table style
	table.SetBorder(false)
	table.SetHeaderLine(true)
	table.SetColumnSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	// Add rows
	for _, item := range result.Items {
		change := FormatDiffFull(item.Diff, item.DiffPct, item.IsNew, item.IsRemoved)
		table.Append([]string{
			Truncate(item.Name, ServiceNameMaxWidth),
			FormatCurrency(item.FromCost),
			FormatCurrency(item.ToCost),
			change,
		})
	}

	table.Render()
	fmt.Fprintln(w)

	return nil
}

// RenderTopTable outputs the top result as a formatted table to stdout
func RenderTopTable(result *diff.TopResult) error {
	return RenderTopTableTo(os.Stdout, result)
}

// RenderTopTableTo outputs the top result as a formatted table to the specified writer
func RenderTopTableTo(w io.Writer, result *diff.TopResult) error {
	// Print header
	fmt.Fprintf(w, "\n%s\n\n", Header(fmt.Sprintf("AWS Top Costs: %s", result.Period.Label())))

	// Print total
	fmt.Fprintf(w, "Total: %s\n\n", FormatCurrency(result.Total))

	if len(result.Items) == 0 {
		fmt.Fprintln(w, Muted("No cost data found for the specified period."))
		return nil
	}

	// Create table
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"#", "Service", "Cost", "% of Total"})

	// Configure table style
	table.SetBorder(false)
	table.SetHeaderLine(true)
	table.SetColumnSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	// Add rows
	for i, item := range result.Items {
		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			Truncate(item.Name, TopServiceNameMaxWidth),
			FormatCurrency(item.Cost),
			fmt.Sprintf("%.1f%%", item.Percent),
		})
	}

	table.Render()
	fmt.Fprintln(w)

	return nil
}

// RenderWatchTable outputs the watch result as a formatted table to stdout
func RenderWatchTable(result *diff.WatchResult) error {
	return RenderWatchTableTo(os.Stdout, result)
}

// RenderWatchTableTo outputs the watch result as a formatted table to the specified writer
func RenderWatchTableTo(w io.Writer, result *diff.WatchResult) error {
	// Print header
	fmt.Fprintf(w, "\n%s\n\n", Header(fmt.Sprintf("AWS Daily Costs: %s to %s",
		result.StartDate.Format("Jan 2"),
		result.EndDate.Format("Jan 2, 2006"))))

	// Print summary
	fmt.Fprintf(w, "Total: %s  |  Daily Average: %s\n\n",
		FormatCurrency(result.Total),
		FormatCurrency(result.Average))

	if len(result.Days) == 0 {
		fmt.Fprintln(w, Muted("No cost data found for the specified period."))
		return nil
	}

	// Create table
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Date", "Day", "Cost", "Change"})

	// Configure table style
	table.SetBorder(false)
	table.SetHeaderLine(true)
	table.SetColumnSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	// Add rows
	for i, day := range result.Days {
		var changeStr string
		if i == 0 {
			changeStr = Muted("-")
		} else {
			changeStr = FormatDiffFull(day.Change, day.ChangePercent, false, false)
		}

		table.Append([]string{
			day.Date.Format("Jan 2"),
			day.Date.Format("Mon"),
			FormatCurrency(day.Cost),
			changeStr,
		})
	}

	table.Render()

	// Print visual bar chart
	fmt.Fprintln(w)
	renderBarChartTo(w, result.Days, result.Average)
	fmt.Fprintln(w)

	return nil
}

// Constants for table display widths
const (
	ServiceNameMaxWidth    = 35 // Max width for service names in diff table
	TopServiceNameMaxWidth = 40 // Max width for service names in top table
	BarChartMaxWidth       = 40 // Max width for ASCII bar chart
	AboveAverageThreshold  = 1.2
	BelowAverageThreshold  = 0.8
)

// renderBarChartTo renders a simple ASCII bar chart to the specified writer
func renderBarChartTo(w io.Writer, days []diff.DayItem, average float64) {
	if len(days) == 0 {
		return
	}

	// Find max cost for scaling
	var maxCost float64
	for _, day := range days {
		if day.Cost > maxCost {
			maxCost = day.Cost
		}
	}

	if maxCost == 0 {
		return
	}

	for _, day := range days {
		width := int((day.Cost / maxCost) * BarChartMaxWidth)
		if width < 1 && day.Cost > 0 {
			width = 1
		}

		bar := strings.Repeat("█", width)

		// Color based on comparison to average
		if day.Cost > average*AboveAverageThreshold {
			bar = Red.Sprint(bar)
		} else if day.Cost < average*BelowAverageThreshold {
			bar = Green.Sprint(bar)
		} else {
			bar = Cyan.Sprint(bar)
		}

		fmt.Fprintf(w, "%s %s %s\n",
			Muted(day.Date.Format("Jan 2")),
			bar,
			Muted(FormatCurrency(day.Cost)))
	}
}

// Truncate shortens a string to maxLen characters, respecting unicode runes
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

