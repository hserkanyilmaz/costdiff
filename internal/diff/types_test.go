package diff

import (
	"testing"
	"time"
)

func TestPeriod_Label(t *testing.T) {
	tests := []struct {
		name   string
		period Period
		want   string
	}{
		{
			name: "full month",
			period: Period{
				Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: "Dec 2024",
		},
		{
			name: "single day",
			period: Period{
				Start: time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 12, 16, 0, 0, 0, 0, time.UTC),
			},
			want: "Dec 15, 2024",
		},
		{
			name: "date range",
			period: Period{
				Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
			},
			want: "Dec 1 - Dec 14, 2024",
		},
		{
			name: "cross-month range",
			period: Period{
				Start: time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			want: "Dec 15 - Jan 14, 2025",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.period.Label(); got != tt.want {
				t.Errorf("Period.Label() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPeriod_ToJSON(t *testing.T) {
	period := Period{
		Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	json := period.ToJSON()

	if json.Start != "2024-12-01" {
		t.Errorf("Start = %q, want %q", json.Start, "2024-12-01")
	}
	if json.End != "2025-01-01" {
		t.Errorf("End = %q, want %q", json.End, "2025-01-01")
	}
	if json.Label != "Dec 2024" {
		t.Errorf("Label = %q, want %q", json.Label, "Dec 2024")
	}
}

func TestResult_ToJSON(t *testing.T) {
	result := &Result{
		FromPeriod: Period{
			Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ToPeriod: Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		FromTotal: 1000,
		ToTotal:   1200,
		TotalDiff: 200,
		TotalPct:  20,
		Items: []Item{
			{Name: "EC2", FromCost: 500, ToCost: 600, Diff: 100, DiffPct: 20},
		},
	}

	json := result.ToJSON()

	if json.FromPeriod.Label != "Dec 2024" {
		t.Errorf("FromPeriod.Label = %q, want %q", json.FromPeriod.Label, "Dec 2024")
	}
	if json.ToPeriod.Label != "Jan 2025" {
		t.Errorf("ToPeriod.Label = %q, want %q", json.ToPeriod.Label, "Jan 2025")
	}
	if json.FromTotal != 1000 {
		t.Errorf("FromTotal = %v, want %v", json.FromTotal, 1000)
	}
	if len(json.Items) != 1 {
		t.Errorf("Items count = %v, want %v", len(json.Items), 1)
	}
}

func TestTopResult_ToJSON(t *testing.T) {
	result := &TopResult{
		Period: Period{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Total: 1000,
		Items: []TopItem{
			{Name: "EC2", Cost: 500, Percent: 50},
			{Name: "S3", Cost: 300, Percent: 30},
		},
	}

	json := result.ToJSON()

	if json.Period.Label != "Jan 2025" {
		t.Errorf("Period.Label = %q, want %q", json.Period.Label, "Jan 2025")
	}
	if json.Total != 1000 {
		t.Errorf("Total = %v, want %v", json.Total, 1000)
	}
	if len(json.Items) != 2 {
		t.Errorf("Items count = %v, want %v", len(json.Items), 2)
	}
}

func TestWatchResult_ToJSON(t *testing.T) {
	result := &WatchResult{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
		Total:     700,
		Average:   100,
		Days: []DayItem{
			{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Cost: 100, Change: 0, ChangePercent: 0},
			{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Cost: 120, Change: 20, ChangePercent: 20},
		},
	}

	json := result.ToJSON()

	if json.StartDate != "2025-01-01" {
		t.Errorf("StartDate = %q, want %q", json.StartDate, "2025-01-01")
	}
	if json.EndDate != "2025-01-08" {
		t.Errorf("EndDate = %q, want %q", json.EndDate, "2025-01-08")
	}
	if json.Total != 700 {
		t.Errorf("Total = %v, want %v", json.Total, 700)
	}
	if json.Average != 100 {
		t.Errorf("Average = %v, want %v", json.Average, 100)
	}
	if len(json.Days) != 2 {
		t.Errorf("Days count = %v, want %v", len(json.Days), 2)
	}
	if json.Days[0].Date != "2025-01-01" {
		t.Errorf("Days[0].Date = %q, want %q", json.Days[0].Date, "2025-01-01")
	}
	if json.Days[1].Change != 20 {
		t.Errorf("Days[1].Change = %v, want %v", json.Days[1].Change, 20)
	}
}

func TestItem_Fields(t *testing.T) {
	item := Item{
		Name:      "EC2",
		FromCost:  100,
		ToCost:    150,
		Diff:      50,
		DiffPct:   50,
		IsNew:     false,
		IsRemoved: false,
	}

	if item.Name != "EC2" {
		t.Errorf("Name = %q, want %q", item.Name, "EC2")
	}
	if item.FromCost != 100 {
		t.Errorf("FromCost = %v, want %v", item.FromCost, 100)
	}
	if item.ToCost != 150 {
		t.Errorf("ToCost = %v, want %v", item.ToCost, 150)
	}
	if item.Diff != 50 {
		t.Errorf("Diff = %v, want %v", item.Diff, 50)
	}
}

func TestDayItem_Fields(t *testing.T) {
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	item := DayItem{
		Date:          date,
		Cost:          150.50,
		Change:        25.50,
		ChangePercent: 20.4,
	}

	if !item.Date.Equal(date) {
		t.Errorf("Date = %v, want %v", item.Date, date)
	}
	if item.Cost != 150.50 {
		t.Errorf("Cost = %v, want %v", item.Cost, 150.50)
	}
	if item.Change != 25.50 {
		t.Errorf("Change = %v, want %v", item.Change, 25.50)
	}
}
