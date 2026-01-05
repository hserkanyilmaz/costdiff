package cmd

import (
	"testing"

	"github.com/hserkanyilmaz/costdiff/internal/aws"
	"github.com/hserkanyilmaz/costdiff/internal/diff"
)

func TestParsePeriods_Defaults(t *testing.T) {
	// Test with empty inputs (defaults)
	from, to, err := parsePeriods("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// From period should start on day 1
	if from.Start.Day() != 1 {
		t.Errorf("from.Start.Day() = %v, want 1", from.Start.Day())
	}

	// To period should start on day 1
	if to.Start.Day() != 1 {
		t.Errorf("to.Start.Day() = %v, want 1", to.Start.Day())
	}

	// From should be before To
	if !from.Start.Before(to.Start) {
		t.Error("from.Start should be before to.Start")
	}
}

func TestParsePeriods_SpecificMonths(t *testing.T) {
	from, to, err := parsePeriods("2024-10", "2024-12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if from.Start.Month() != 10 || from.Start.Year() != 2024 {
		t.Errorf("from.Start = %v, want October 2024", from.Start)
	}
	if to.Start.Month() != 12 || to.Start.Year() != 2024 {
		t.Errorf("to.Start = %v, want December 2024", to.Start)
	}
}

func TestParsePeriods_SpecificDays(t *testing.T) {
	from, to, err := parsePeriods("2024-10-15", "2024-12-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if from.Start.Day() != 15 {
		t.Errorf("from.Start.Day() = %v, want 15", from.Start.Day())
	}
	if to.Start.Day() != 20 {
		t.Errorf("to.Start.Day() = %v, want 20", to.Start.Day())
	}
}

func TestParsePeriods_InvalidFrom(t *testing.T) {
	_, _, err := parsePeriods("invalid", "2024-12")
	if err == nil {
		t.Error("expected error for invalid from date")
	}
}

func TestParsePeriods_InvalidTo(t *testing.T) {
	_, _, err := parsePeriods("2024-10", "invalid")
	if err == nil {
		t.Error("expected error for invalid to date")
	}
}

func TestParsePeriods_FromAfterTo(t *testing.T) {
	_, _, err := parsePeriods("2024-12", "2024-10")
	if err == nil {
		t.Error("expected error when from date is after to date")
	}
}

func TestParsePeriods_SameDates(t *testing.T) {
	_, _, err := parsePeriods("2024-10", "2024-10")
	if err == nil {
		t.Error("expected error when from and to dates are the same")
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input     string
		wantStart string
		wantEnd   string
		wantErr   bool
	}{
		{
			input:     "2024-12",
			wantStart: "2024-12-01",
			wantEnd:   "2025-01-01",
		},
		{
			input:     "2024-12-15",
			wantStart: "2024-12-15",
			wantEnd:   "2024-12-16",
		},
		{
			input:     "2024-01",
			wantStart: "2024-01-01",
			wantEnd:   "2024-02-01",
		},
		{
			input:   "invalid",
			wantErr: true,
		},
		{
			input:   "2024",
			wantErr: true,
		},
		{
			input:   "12-2024",
			wantErr: true,
		},
		{
			input:   "2024/12/15",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			period, err := parseDate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotStart := period.Start.Format("2006-01-02")
			gotEnd := period.End.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("Start = %q, want %q", gotStart, tt.wantStart)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("End = %q, want %q", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestParseGroupBy(t *testing.T) {
	tests := []struct {
		group    string
		tag      string
		wantType string
		wantKey  string
		wantErr  bool
	}{
		{group: "service", wantType: "DIMENSION", wantKey: "SERVICE"},
		{group: "region", wantType: "DIMENSION", wantKey: "REGION"},
		{group: "account", wantType: "DIMENSION", wantKey: "LINKED_ACCOUNT"},
		{group: "tag", tag: "team", wantType: "TAG", wantKey: "team"},
		{group: "tag", tag: "environment", wantType: "TAG", wantKey: "environment"},
		{group: "tag", tag: "", wantErr: true},
		{group: "invalid", wantErr: true},
		{group: "SERVICE", wantErr: true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.group+"_"+tt.tag, func(t *testing.T) {
			got, err := parseGroupBy(tt.group, tt.tag)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Key != tt.wantKey {
				t.Errorf("Key = %q, want %q", got.Key, tt.wantKey)
			}
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
		})
	}
}

func TestFilterByThreshold(t *testing.T) {
	result := &diff.Result{
		FromPeriod: diff.Period{},
		ToPeriod:   diff.Period{},
		FromTotal:  300,
		ToTotal:    350,
		Items: []diff.Item{
			{Name: "A", Diff: 100},
			{Name: "B", Diff: -50},
			{Name: "C", Diff: 10},
			{Name: "D", Diff: -5},
		},
	}

	tests := []struct {
		threshold float64
		wantCount int
	}{
		{0, 4},
		{5, 4},
		{10, 3},
		{20, 2},
		{50, 2},
		{100, 1},
		{200, 0},
	}

	for _, tt := range tests {
		filtered := filterByThreshold(result, tt.threshold)
		if len(filtered.Items) != tt.wantCount {
			t.Errorf("threshold=%v: got %d items, want %d", tt.threshold, len(filtered.Items), tt.wantCount)
		}
	}
}

func TestFilterByThreshold_PreservesPeriods(t *testing.T) {
	result := &diff.Result{
		FromTotal: 100,
		ToTotal:   150,
		Items: []diff.Item{
			{Name: "A", Diff: 50},
		},
	}

	filtered := filterByThreshold(result, 10)

	if filtered.FromTotal != result.FromTotal {
		t.Errorf("FromTotal not preserved: got %v, want %v", filtered.FromTotal, result.FromTotal)
	}
	if filtered.ToTotal != result.ToTotal {
		t.Errorf("ToTotal not preserved: got %v, want %v", filtered.ToTotal, result.ToTotal)
	}
}

func TestGetAWSMetric(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{input: "net-amortized", want: "NetAmortizedCost"},
		{input: "amortized", want: "AmortizedCost"},
		{input: "unblended", want: "UnblendedCost"},
		{input: "blended", want: "BlendedCost"},
		{input: "net-unblended", want: "NetUnblendedCost"},
		{input: "normalized", want: "NormalizedUsageAmount"},
		{input: "usage-quantity", want: "UsageQuantity"},
		{input: "invalid", wantErr: true},
		{input: "NetAmortizedCost", wantErr: true}, // Full AWS name not accepted
		{input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			costMetric = tt.input
			got, err := getAWSMetric()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("getAWSMetric() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAbsFunction(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-0.5, 0.5},
		{0.5, 0.5},
		{-1000000, 1000000},
	}

	for _, tt := range tests {
		if got := diff.Abs(tt.input); got != tt.want {
			t.Errorf("diff.Abs(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}


func TestOutputResult_InvalidFormat(t *testing.T) {
	result := &diff.Result{}
	err := outputResult(result, "invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestGroupTypeConstants(t *testing.T) {
	// Verify the predefined group types
	if aws.GroupByService.Type != "DIMENSION" || aws.GroupByService.Key != "SERVICE" {
		t.Errorf("GroupByService = %+v, want DIMENSION/SERVICE", aws.GroupByService)
	}
	if aws.GroupByRegion.Type != "DIMENSION" || aws.GroupByRegion.Key != "REGION" {
		t.Errorf("GroupByRegion = %+v, want DIMENSION/REGION", aws.GroupByRegion)
	}
	if aws.GroupByAccount.Type != "DIMENSION" || aws.GroupByAccount.Key != "LINKED_ACCOUNT" {
		t.Errorf("GroupByAccount = %+v, want DIMENSION/LINKED_ACCOUNT", aws.GroupByAccount)
	}
}

