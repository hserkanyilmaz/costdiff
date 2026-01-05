package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/hserkanyilmaz/costdiff/internal/diff"
)

func TestRenderJSONTo(t *testing.T) {
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
	err := RenderJSONTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderJSONTo() error = %v", err)
	}

	// Parse the output to verify it's valid JSON
	var output diff.ResultJSON
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify the values
	if output.FromTotal != 1000 {
		t.Errorf("FromTotal = %v, want %v", output.FromTotal, 1000)
	}
	if output.ToTotal != 1200 {
		t.Errorf("ToTotal = %v, want %v", output.ToTotal, 1200)
	}
	if output.TotalDiff != 200 {
		t.Errorf("TotalDiff = %v, want %v", output.TotalDiff, 200)
	}
	if len(output.Items) != 2 {
		t.Errorf("Items count = %v, want %v", len(output.Items), 2)
	}
	if output.FromPeriod.Start != "2024-12-01" {
		t.Errorf("FromPeriod.Start = %v, want %v", output.FromPeriod.Start, "2024-12-01")
	}
	if output.ToPeriod.Label != "Jan 2025" {
		t.Errorf("ToPeriod.Label = %v, want %v", output.ToPeriod.Label, "Jan 2025")
	}
}

func TestRenderJSONTo_EmptyItems(t *testing.T) {
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
	err := RenderJSONTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderJSONTo() error = %v", err)
	}

	// Parse the output to verify it's valid JSON
	var output diff.ResultJSON
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(output.Items) != 0 {
		t.Errorf("Items count = %v, want %v", len(output.Items), 0)
	}
}

func TestRenderTopJSONTo(t *testing.T) {
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
	err := RenderTopJSONTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderTopJSONTo() error = %v", err)
	}

	// Parse the output to verify it's valid JSON
	var output diff.TopResultJSON
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify the values
	if output.Total != 1000 {
		t.Errorf("Total = %v, want %v", output.Total, 1000)
	}
	if len(output.Items) != 2 {
		t.Errorf("Items count = %v, want %v", len(output.Items), 2)
	}
	if output.Items[0].Name != "EC2" {
		t.Errorf("Items[0].Name = %v, want %v", output.Items[0].Name, "EC2")
	}
	if output.Items[0].Percent != 50 {
		t.Errorf("Items[0].Percent = %v, want %v", output.Items[0].Percent, 50)
	}
	if output.Period.Label != "Jan 2025" {
		t.Errorf("Period.Label = %v, want %v", output.Period.Label, "Jan 2025")
	}
}

func TestRenderWatchJSONTo(t *testing.T) {
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
	err := RenderWatchJSONTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderWatchJSONTo() error = %v", err)
	}

	// Parse the output to verify it's valid JSON
	var output diff.WatchResultJSON
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify the values
	if output.Total != 700 {
		t.Errorf("Total = %v, want %v", output.Total, 700)
	}
	if output.Average != 100 {
		t.Errorf("Average = %v, want %v", output.Average, 100)
	}
	if output.StartDate != "2025-01-01" {
		t.Errorf("StartDate = %v, want %v", output.StartDate, "2025-01-01")
	}
	if output.EndDate != "2025-01-08" {
		t.Errorf("EndDate = %v, want %v", output.EndDate, "2025-01-08")
	}
	if len(output.Days) != 2 {
		t.Errorf("Days count = %v, want %v", len(output.Days), 2)
	}
	if output.Days[0].Date != "2025-01-01" {
		t.Errorf("Days[0].Date = %v, want %v", output.Days[0].Date, "2025-01-01")
	}
	if output.Days[1].Change != 20 {
		t.Errorf("Days[1].Change = %v, want %v", output.Days[1].Change, 20)
	}
}

func TestRenderJSONTo_Indentation(t *testing.T) {
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
	err := RenderJSONTo(&buf, result)
	if err != nil {
		t.Fatalf("RenderJSONTo() error = %v", err)
	}

	// Check that output is indented (contains newlines and spaces)
	output := buf.String()
	if !containsIndentation(output) {
		t.Error("JSON output should be indented")
	}
}

func containsIndentation(s string) bool {
	// Check for newlines followed by spaces (indentation)
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '\n' && i+1 < len(s) && s[i+1] == ' ' {
			return true
		}
	}
	return false
}

