package output

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/hserkanyilmaz/costdiff/internal/diff"
)

func TestRenderCSVTo(t *testing.T) {
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
			{Name: "EC2", FromCost: 500, ToCost: 600, Diff: 100, DiffPct: 20, IsNew: false, IsRemoved: false},
			{Name: "S3", FromCost: 300, ToCost: 350, Diff: 50, DiffPct: 16.67, IsNew: false, IsRemoved: false},
		},
	}

	var buf bytes.Buffer
	err := RenderCSVTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderCSVTo() error = %v", err)
	}

	// Parse the CSV output
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV output: %v", err)
	}

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("CSV records count = %v, want %v", len(records), 3)
	}

	// Verify header
	expectedHeader := []string{"name", "from_period", "to_period", "from_cost", "to_cost", "diff", "diff_percent", "is_new", "is_removed"}
	if len(records[0]) != len(expectedHeader) {
		t.Errorf("Header column count = %v, want %v", len(records[0]), len(expectedHeader))
	}
	for i, col := range expectedHeader {
		if records[0][i] != col {
			t.Errorf("Header[%d] = %v, want %v", i, records[0][i], col)
		}
	}

	// Verify first data row (EC2)
	if records[1][0] != "EC2" {
		t.Errorf("Row 1 name = %v, want %v", records[1][0], "EC2")
	}
	if records[1][1] != "Dec 2024" {
		t.Errorf("Row 1 from_period = %v, want %v", records[1][1], "Dec 2024")
	}
	if records[1][2] != "Jan 2025" {
		t.Errorf("Row 1 to_period = %v, want %v", records[1][2], "Jan 2025")
	}
	if records[1][3] != "500.00" {
		t.Errorf("Row 1 from_cost = %v, want %v", records[1][3], "500.00")
	}
	if records[1][4] != "600.00" {
		t.Errorf("Row 1 to_cost = %v, want %v", records[1][4], "600.00")
	}
}

func TestRenderCSVTo_EmptyItems(t *testing.T) {
	result := &diff.Result{
		FromPeriod: diff.Period{
			Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ToPeriod: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Items: []diff.Item{},
	}

	var buf bytes.Buffer
	err := RenderCSVTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderCSVTo() error = %v", err)
	}

	// Parse the CSV output
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV output: %v", err)
	}

	// Should have only header row
	if len(records) != 1 {
		t.Errorf("CSV records count = %v, want %v (header only)", len(records), 1)
	}
}

func TestRenderCSVTo_NewAndRemovedItems(t *testing.T) {
	result := &diff.Result{
		FromPeriod: diff.Period{
			Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ToPeriod: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Items: []diff.Item{
			{Name: "NewService", FromCost: 0, ToCost: 200, Diff: 200, DiffPct: 100, IsNew: true, IsRemoved: false},
			{Name: "RemovedService", FromCost: 100, ToCost: 0, Diff: -100, DiffPct: -100, IsNew: false, IsRemoved: true},
		},
	}

	var buf bytes.Buffer
	err := RenderCSVTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderCSVTo() error = %v", err)
	}

	// Parse the CSV output
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV output: %v", err)
	}

	// Verify is_new and is_removed columns
	if records[1][7] != "true" { // is_new for NewService
		t.Errorf("Row 1 is_new = %v, want %v", records[1][7], "true")
	}
	if records[1][8] != "false" { // is_removed for NewService
		t.Errorf("Row 1 is_removed = %v, want %v", records[1][8], "false")
	}
	if records[2][7] != "false" { // is_new for RemovedService
		t.Errorf("Row 2 is_new = %v, want %v", records[2][7], "false")
	}
	if records[2][8] != "true" { // is_removed for RemovedService
		t.Errorf("Row 2 is_removed = %v, want %v", records[2][8], "true")
	}
}

func TestRenderTopCSVTo(t *testing.T) {
	result := &diff.TopResult{
		Period: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Total: 1000,
		Items: []diff.TopItem{
			{Name: "EC2", Cost: 500, Percent: 50},
			{Name: "S3", Cost: 300, Percent: 30},
		},
	}

	var buf bytes.Buffer
	err := RenderTopCSVTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderTopCSVTo() error = %v", err)
	}

	// Parse the CSV output
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV output: %v", err)
	}

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("CSV records count = %v, want %v", len(records), 3)
	}

	// Verify header
	expectedHeader := []string{"rank", "name", "period", "cost", "percent"}
	for i, col := range expectedHeader {
		if records[0][i] != col {
			t.Errorf("Header[%d] = %v, want %v", i, records[0][i], col)
		}
	}

	// Verify first data row
	if records[1][0] != "1" {
		t.Errorf("Row 1 rank = %v, want %v", records[1][0], "1")
	}
	if records[1][1] != "EC2" {
		t.Errorf("Row 1 name = %v, want %v", records[1][1], "EC2")
	}
	if records[1][2] != "Jan 2025" {
		t.Errorf("Row 1 period = %v, want %v", records[1][2], "Jan 2025")
	}
	if records[1][3] != "500.00" {
		t.Errorf("Row 1 cost = %v, want %v", records[1][3], "500.00")
	}
	if records[1][4] != "50.00" {
		t.Errorf("Row 1 percent = %v, want %v", records[1][4], "50.00")
	}
}

func TestRenderWatchCSVTo(t *testing.T) {
	result := &diff.WatchResult{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
		Total:     700,
		Average:   100,
		Days: []diff.DayItem{
			{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Cost: 100, Change: 0, ChangePercent: 0},
			{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Cost: 120, Change: 20, ChangePercent: 20},
		},
	}

	var buf bytes.Buffer
	err := RenderWatchCSVTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderWatchCSVTo() error = %v", err)
	}

	// Parse the CSV output
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV output: %v", err)
	}

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("CSV records count = %v, want %v", len(records), 3)
	}

	// Verify header
	expectedHeader := []string{"date", "day_of_week", "cost", "change", "change_percent"}
	for i, col := range expectedHeader {
		if records[0][i] != col {
			t.Errorf("Header[%d] = %v, want %v", i, records[0][i], col)
		}
	}

	// Verify first data row
	if records[1][0] != "2025-01-01" {
		t.Errorf("Row 1 date = %v, want %v", records[1][0], "2025-01-01")
	}
	if records[1][1] != "Wednesday" {
		t.Errorf("Row 1 day_of_week = %v, want %v", records[1][1], "Wednesday")
	}
	if records[1][2] != "100.00" {
		t.Errorf("Row 1 cost = %v, want %v", records[1][2], "100.00")
	}

	// Verify second data row has change
	if records[2][3] != "20.00" {
		t.Errorf("Row 2 change = %v, want %v", records[2][3], "20.00")
	}
	if records[2][4] != "20.00" {
		t.Errorf("Row 2 change_percent = %v, want %v", records[2][4], "20.00")
	}
}

func TestRenderTopCSVTo_EmptyItems(t *testing.T) {
	result := &diff.TopResult{
		Period: diff.Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Total: 0,
		Items: []diff.TopItem{},
	}

	var buf bytes.Buffer
	err := RenderTopCSVTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderTopCSVTo() error = %v", err)
	}

	// Parse the CSV output
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV output: %v", err)
	}

	// Should have only header row
	if len(records) != 1 {
		t.Errorf("CSV records count = %v, want %v (header only)", len(records), 1)
	}
}

func TestRenderWatchCSVTo_EmptyDays(t *testing.T) {
	result := &diff.WatchResult{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
		Total:     0,
		Average:   0,
		Days:      []diff.DayItem{},
	}

	var buf bytes.Buffer
	err := RenderWatchCSVTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderWatchCSVTo() error = %v", err)
	}

	// Parse the CSV output
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV output: %v", err)
	}

	// Should have only header row
	if len(records) != 1 {
		t.Errorf("CSV records count = %v, want %v (header only)", len(records), 1)
	}
}

