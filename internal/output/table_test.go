package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/hserkanyilmaz/costdiff/internal/diff"
)

func init() {
	// Disable colors for testing
	color.NoColor = true
}

func TestRenderTableTo(t *testing.T) {
	result := &diff.Result{
		FromPeriod: diff.Period{
			Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ToPeriod: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		FromTotal: 1000,
		ToTotal:   1200,
		TotalDiff: 200,
		TotalPct:  20,
		Items: []diff.Item{
			{Name: "EC2", FromCost: 500, ToCost: 600, Diff: 100, DiffPct: 20},
			{Name: "S3", FromCost: 300, ToCost: 350, Diff: 50, DiffPct: 16.67},
		},
	}

	var buf bytes.Buffer
	err := RenderTableTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderTableTo() error = %v", err)
	}

	output := buf.String()

	// Check that the header is present
	if !strings.Contains(output, "AWS Cost Diff") {
		t.Error("Output should contain 'AWS Cost Diff' header")
	}

	// Check that period labels are present
	if !strings.Contains(output, "Dec 2024") {
		t.Error("Output should contain 'Dec 2024'")
	}
	if !strings.Contains(output, "Jan 2025") {
		t.Error("Output should contain 'Jan 2025'")
	}

	// Check that totals are present
	if !strings.Contains(output, "$1000.00") {
		t.Error("Output should contain from total '$1000.00'")
	}
	if !strings.Contains(output, "$1200.00") {
		t.Error("Output should contain to total '$1200.00'")
	}

	// Check that items are present
	if !strings.Contains(output, "EC2") {
		t.Error("Output should contain 'EC2'")
	}
	if !strings.Contains(output, "S3") {
		t.Error("Output should contain 'S3'")
	}
}

func TestRenderTableTo_EmptyItems(t *testing.T) {
	result := &diff.Result{
		FromPeriod: diff.Period{
			Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ToPeriod: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		FromTotal: 0,
		ToTotal:   0,
		TotalDiff: 0,
		TotalPct:  0,
		Items:     []diff.Item{},
	}

	var buf bytes.Buffer
	err := RenderTableTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderTableTo() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No cost data found") {
		t.Error("Output should contain 'No cost data found' message")
	}
}

func TestRenderTopTableTo(t *testing.T) {
	result := &diff.TopResult{
		Period: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Total: 1000,
		Items: []diff.TopItem{
			{Name: "EC2", Cost: 500, Percent: 50},
			{Name: "S3", Cost: 300, Percent: 30},
			{Name: "RDS", Cost: 200, Percent: 20},
		},
	}

	var buf bytes.Buffer
	err := RenderTopTableTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderTopTableTo() error = %v", err)
	}

	output := buf.String()

	// Check that the header is present
	if !strings.Contains(output, "AWS Top Costs") {
		t.Error("Output should contain 'AWS Top Costs' header")
	}

	// Check that period is present
	if !strings.Contains(output, "Jan 2025") {
		t.Error("Output should contain 'Jan 2025'")
	}

	// Check that total is present
	if !strings.Contains(output, "$1000.00") {
		t.Error("Output should contain total '$1000.00'")
	}

	// Check that items are present with percentages
	if !strings.Contains(output, "EC2") {
		t.Error("Output should contain 'EC2'")
	}
	if !strings.Contains(output, "50.0%") {
		t.Error("Output should contain '50.0%'")
	}
}

func TestRenderTopTableTo_EmptyItems(t *testing.T) {
	result := &diff.TopResult{
		Period: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Total: 0,
		Items: []diff.TopItem{},
	}

	var buf bytes.Buffer
	err := RenderTopTableTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderTopTableTo() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No cost data found") {
		t.Error("Output should contain 'No cost data found' message")
	}
}

func TestRenderWatchTableTo(t *testing.T) {
	result := &diff.WatchResult{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
		Total:     700,
		Average:   100,
		Days: []diff.DayItem{
			{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Cost: 100, Change: 0, ChangePercent: 0},
			{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Cost: 120, Change: 20, ChangePercent: 20},
			{Date: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), Cost: 80, Change: -40, ChangePercent: -33.33},
		},
	}

	var buf bytes.Buffer
	err := RenderWatchTableTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderWatchTableTo() error = %v", err)
	}

	output := buf.String()

	// Check that the header is present
	if !strings.Contains(output, "AWS Daily Costs") {
		t.Error("Output should contain 'AWS Daily Costs' header")
	}

	// Check that total and average are present
	if !strings.Contains(output, "$700.00") {
		t.Error("Output should contain total '$700.00'")
	}
	if !strings.Contains(output, "$100.00") {
		t.Error("Output should contain average '$100.00'")
	}

	// Check that dates are present
	if !strings.Contains(output, "Jan 1") {
		t.Error("Output should contain 'Jan 1'")
	}
	if !strings.Contains(output, "Jan 2") {
		t.Error("Output should contain 'Jan 2'")
	}
}

func TestRenderWatchTableTo_EmptyDays(t *testing.T) {
	result := &diff.WatchResult{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
		Total:     0,
		Average:   0,
		Days:      []diff.DayItem{},
	}

	var buf bytes.Buffer
	err := RenderWatchTableTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderWatchTableTo() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No cost data found") {
		t.Error("Output should contain 'No cost data found' message")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input   string
		maxLen  int
		want    string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"ab", 3, "ab"},
		{"unicode: 日本語", 12, "unicode: 日本語"}, // 12 runes exactly, no truncation
		{"unicode: 日本語テスト", 12, "unicode: ..."},   // 15 runes, truncated
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestRenderTableTo_NewAndRemovedItems(t *testing.T) {
	result := &diff.Result{
		FromPeriod: diff.Period{
			Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ToPeriod: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		FromTotal: 500,
		ToTotal:   600,
		TotalDiff: 100,
		TotalPct:  20,
		Items: []diff.Item{
			{Name: "NewService", FromCost: 0, ToCost: 200, Diff: 200, DiffPct: 100, IsNew: true},
			{Name: "RemovedService", FromCost: 100, ToCost: 0, Diff: -100, DiffPct: -100, IsRemoved: true},
		},
	}

	var buf bytes.Buffer
	err := RenderTableTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderTableTo() error = %v", err)
	}

	output := buf.String()

	// Check that new and removed items are present
	if !strings.Contains(output, "NewService") {
		t.Error("Output should contain 'NewService'")
	}
	if !strings.Contains(output, "RemovedService") {
		t.Error("Output should contain 'RemovedService'")
	}
	if !strings.Contains(output, "new") {
		t.Error("Output should contain 'new' label")
	}
	if !strings.Contains(output, "removed") {
		t.Error("Output should contain 'removed' label")
	}
}

