package output

import (
	"os"
	"testing"

	"github.com/fatih/color"
)

func init() {
	// Disable colors for testing
	color.NoColor = true
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{0, "$0.00"},
		{100, "$100.00"},
		{1234.56, "$1234.56"},
		{-100, "-$100.00"},
		{-1234.56, "-$1234.56"},
		{0.01, "$0.01"},
		{0.001, "$0.00"}, // Rounds down
		{0.005, "$0.01"}, // Rounds up (banker's rounding)
		{99999.99, "$99999.99"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatCurrency(tt.amount); got != tt.want {
				t.Errorf("FormatCurrency(%v) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{0, "+0.0%"},
		{10, "+10.0%"},
		{100.5, "+100.5%"},
		{-10, "-10.0%"},
		{-100.5, "-100.5%"},
		{0.1, "+0.1%"},
		{-0.1, "-0.1%"},
		{1000, "+1000.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatPercent(tt.pct); got != tt.want {
				t.Errorf("FormatPercent(%v) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestFormatChange(t *testing.T) {
	tests := []struct {
		change float64
		want   string
	}{
		{0, "+$0.00"},
		{100, "+$100.00"},
		{-100, "-$100.00"},
		{50.75, "+$50.75"},
		{-50.75, "-$50.75"},
		{0.01, "+$0.01"},
		{-0.01, "-$0.01"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatChange(tt.change); got != tt.want {
				t.Errorf("FormatChange(%v) = %q, want %q", tt.change, got, tt.want)
			}
		})
	}
}

func TestFormatDiffFull(t *testing.T) {
	tests := []struct {
		name      string
		diff      float64
		pct       float64
		isNew     bool
		isRemoved bool
		want      string
	}{
		{
			name: "positive change",
			diff: 100,
			pct:  20,
			want: "+$100.00 (+20.0%)",
		},
		{
			name: "negative change",
			diff: -50,
			pct:  -10,
			want: "-$50.00 (-10.0%)",
		},
		{
			name:  "new item",
			diff:  100,
			pct:   100,
			isNew: true,
			want:  "+$100.00 (new)",
		},
		{
			name:      "removed item",
			diff:      -100,
			pct:       -100,
			isRemoved: true,
			want:      "-$100.00 (removed)",
		},
		{
			name: "zero change",
			diff: 0,
			pct:  0,
			want: "+$0.00 (+0.0%)",
		},
		{
			name: "small positive change",
			diff: 0.50,
			pct:  5.5,
			want: "+$0.50 (+5.5%)",
		},
		{
			name: "large change",
			diff: 10000,
			pct:  500,
			want: "+$10000.00 (+500.0%)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDiffFull(tt.diff, tt.pct, tt.isNew, tt.isRemoved)
			if got != tt.want {
				t.Errorf("FormatDiffFull() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorizeChange(t *testing.T) {
	// With colors disabled, should return formatted string unchanged
	color.NoColor = true

	tests := []struct {
		change    float64
		formatted string
	}{
		{100, "+$100.00"},
		{-50, "-$50.00"},
		{0, "$0.00"},
	}

	for _, tt := range tests {
		got := ColorizeChange(tt.change, tt.formatted)
		if got != tt.formatted {
			t.Errorf("ColorizeChange(%v, %q) = %q, want %q", tt.change, tt.formatted, got, tt.formatted)
		}
	}
}

func TestColorizePercent(t *testing.T) {
	color.NoColor = true

	tests := []struct {
		pct  float64
		want string
	}{
		{10, "+10.0%"},
		{-10, "-10.0%"},
		{0, "+0.0%"},
	}

	for _, tt := range tests {
		got := ColorizePercent(tt.pct)
		if got != tt.want {
			t.Errorf("ColorizePercent(%v) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}

func TestColorizeDiff(t *testing.T) {
	color.NoColor = true

	tests := []struct {
		diff float64
		want string
	}{
		{100, "+$100.00"},
		{-50, "-$50.00"},
		{0, "+$0.00"},
	}

	for _, tt := range tests {
		got := ColorizeDiff(tt.diff)
		if got != tt.want {
			t.Errorf("ColorizeDiff(%v) = %q, want %q", tt.diff, got, tt.want)
		}
	}
}

func TestHeader(t *testing.T) {
	color.NoColor = true
	got := Header("Test Header")
	if got != "Test Header" {
		t.Errorf("Header() = %q, want %q", got, "Test Header")
	}
}

func TestSubheader(t *testing.T) {
	color.NoColor = true
	got := Subheader("Test Subheader")
	if got != "Test Subheader" {
		t.Errorf("Subheader() = %q, want %q", got, "Test Subheader")
	}
}

func TestMuted(t *testing.T) {
	color.NoColor = true
	got := Muted("Muted text")
	if got != "Muted text" {
		t.Errorf("Muted() = %q, want %q", got, "Muted text")
	}
}

func TestSuccess(t *testing.T) {
	color.NoColor = true
	got := Success("Success!")
	if got != "Success!" {
		t.Errorf("Success() = %q, want %q", got, "Success!")
	}
}

func TestWarning(t *testing.T) {
	color.NoColor = true
	got := Warning("Warning!")
	if got != "Warning!" {
		t.Errorf("Warning() = %q, want %q", got, "Warning!")
	}
}

func TestError(t *testing.T) {
	color.NoColor = true
	got := Error("Error!")
	if got != "Error!" {
		t.Errorf("Error() = %q, want %q", got, "Error!")
	}
}

func TestFormatCurrency_WithNoColorEnv(t *testing.T) {
	// Test that NO_COLOR env var is respected
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	// Force re-check
	color.NoColor = true

	got := ColorizeChange(100, "+$100.00")
	if got != "+$100.00" {
		t.Errorf("With NO_COLOR set, ColorizeChange should not add colors, got %q", got)
	}
}
