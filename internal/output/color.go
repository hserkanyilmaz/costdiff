package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	// Colors for output
	Red    = color.New(color.FgRed)
	Green  = color.New(color.FgGreen)
	Yellow = color.New(color.FgYellow)
	Cyan   = color.New(color.FgCyan)
	Bold   = color.New(color.Bold)
	Dim    = color.New(color.Faint)
)

func init() {
	// Disable colors if not a terminal or if NO_COLOR is set
	if os.Getenv("NO_COLOR") != "" || !isTerminal() {
		color.NoColor = true
	}
}

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// ColorizeChange returns a colored string based on whether the change is positive or negative
func ColorizeChange(change float64, formatted string) string {
	if color.NoColor {
		return formatted
	}
	if change > 0 {
		return Red.Sprint(formatted)
	} else if change < 0 {
		return Green.Sprint(formatted)
	}
	return formatted
}

// ColorizePercent returns a colored percentage string
func ColorizePercent(pct float64) string {
	formatted := FormatPercent(pct)
	return ColorizeChange(pct, formatted)
}

// ColorizeDiff returns a colored diff string
func ColorizeDiff(diff float64) string {
	formatted := FormatChange(diff)
	return ColorizeChange(diff, formatted)
}

// FormatCurrency formats a float as a dollar amount
func FormatCurrency(amount float64) string {
	if amount < 0 {
		return fmt.Sprintf("-$%.2f", -amount)
	}
	return fmt.Sprintf("$%.2f", amount)
}

// FormatPercent formats a float as a percentage with sign
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

// FormatDiffFull formats a complete diff string with change and percentage
func FormatDiffFull(diff, pct float64, isNew, isRemoved bool) string {
	if isNew {
		return ColorizeChange(diff, fmt.Sprintf("+$%.2f (new)", diff))
	}
	if isRemoved {
		return ColorizeChange(diff, fmt.Sprintf("-$%.2f (removed)", -diff))
	}
	return ColorizeChange(diff, fmt.Sprintf("%s (%s)", FormatChange(diff), FormatPercent(pct)))
}

// Header prints a styled header
func Header(text string) string {
	return Bold.Sprint(text)
}

// Subheader prints a styled subheader
func Subheader(text string) string {
	return Cyan.Sprint(text)
}

// Muted prints muted/dim text
func Muted(text string) string {
	return Dim.Sprint(text)
}

// Success prints green text
func Success(text string) string {
	return Green.Sprint(text)
}

// Warning prints yellow text
func Warning(text string) string {
	return Yellow.Sprint(text)
}

// Error prints red text
func Error(text string) string {
	return Red.Sprint(text)
}
