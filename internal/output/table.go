package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/hsy/costdiff/internal/diff"
)

// RenderTable outputs the diff result as a formatted table
func RenderTable(result *diff.Result) error {
	// Print header
	fmt.Printf("\n%s\n\n", Header(fmt.Sprintf("AWS Cost Diff: %s → %s",
		result.FromPeriod.Label(),
		result.ToPeriod.Label())))

	// Print total
	totalChange := FormatDiffFull(result.TotalDiff, result.TotalPct, false, false)
	fmt.Printf("Total: %s → %s (%s)\n\n",
		FormatCurrency(result.FromTotal),
		FormatCurrency(result.ToTotal),
		totalChange)

	if len(result.Items) == 0 {
		fmt.Println(Muted("No cost data found for the specified period."))
		return nil
	}

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
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
			truncate(item.Name, 35),
			FormatCurrency(item.FromCost),
			FormatCurrency(item.ToCost),
			change,
		})
	}

	table.Render()
	fmt.Println()

	return nil
}

// RenderTopTable outputs the top result as a formatted table
func RenderTopTable(result *diff.TopResult) error {
	// Print header
	fmt.Printf("\n%s\n\n", Header(fmt.Sprintf("AWS Top Costs: %s", result.Period.Label())))

	// Print total
	fmt.Printf("Total: %s\n\n", FormatCurrency(result.Total))

	if len(result.Items) == 0 {
		fmt.Println(Muted("No cost data found for the specified period."))
		return nil
	}

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
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
			truncate(item.Name, 40),
			FormatCurrency(item.Cost),
			fmt.Sprintf("%.1f%%", item.Percent),
		})
	}

	table.Render()
	fmt.Println()

	return nil
}

// RenderWatchTable outputs the watch result as a formatted table
func RenderWatchTable(result *diff.WatchResult) error {
	// Print header
	fmt.Printf("\n%s\n\n", Header(fmt.Sprintf("AWS Daily Costs: %s to %s",
		result.StartDate.Format("Jan 2"),
		result.EndDate.Format("Jan 2, 2006"))))

	// Print summary
	fmt.Printf("Total: %s  |  Daily Average: %s\n\n",
		FormatCurrency(result.Total),
		FormatCurrency(result.Average))

	if len(result.Days) == 0 {
		fmt.Println(Muted("No cost data found for the specified period."))
		return nil
	}

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
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
	fmt.Println()
	renderBarChart(result.Days, result.Average)
	fmt.Println()

	return nil
}

// renderBarChart renders a simple ASCII bar chart
func renderBarChart(days []diff.DayItem, average float64) {
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

	const maxWidth = 40

	for _, day := range days {
		width := int((day.Cost / maxCost) * maxWidth)
		if width < 1 && day.Cost > 0 {
			width = 1
		}

		bar := strings.Repeat("█", width)

		// Color based on comparison to average
		if day.Cost > average*1.2 {
			bar = Red.Sprint(bar)
		} else if day.Cost < average*0.8 {
			bar = Green.Sprint(bar)
		} else {
			bar = Cyan.Sprint(bar)
		}

		fmt.Printf("%s %s %s\n",
			Muted(day.Date.Format("Jan 2")),
			bar,
			Muted(FormatCurrency(day.Cost)))
	}
}

// truncate shortens a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

