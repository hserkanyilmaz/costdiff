package diff

import (
	"testing"
	"time"
)

func TestCompare(t *testing.T) {
	fromPeriod := Period{
		Start: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	toPeriod := Period{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name       string
		fromCosts  map[string]float64
		toCosts    map[string]float64
		wantTotal  float64
		wantDiff   float64
		wantPct    float64
		wantItems  int
	}{
		{
			name:       "basic comparison",
			fromCosts:  map[string]float64{"EC2": 100, "S3": 50},
			toCosts:    map[string]float64{"EC2": 120, "S3": 40},
			wantTotal:  160,
			wantDiff:   10,
			wantPct:    6.666666666666667,
			wantItems:  2,
		},
		{
			name:       "empty from costs",
			fromCosts:  map[string]float64{},
			toCosts:    map[string]float64{"EC2": 100},
			wantTotal:  100,
			wantDiff:   100,
			wantPct:    0, // Division by zero case
			wantItems:  1,
		},
		{
			name:       "empty to costs",
			fromCosts:  map[string]float64{"EC2": 100},
			toCosts:    map[string]float64{},
			wantTotal:  0,
			wantDiff:   -100,
			wantPct:    -100,
			wantItems:  1,
		},
		{
			name:       "new service added",
			fromCosts:  map[string]float64{"EC2": 100},
			toCosts:    map[string]float64{"EC2": 100, "Lambda": 50},
			wantTotal:  150,
			wantDiff:   50,
			wantPct:    50,
			wantItems:  2,
		},
		{
			name:       "service removed",
			fromCosts:  map[string]float64{"EC2": 100, "Lambda": 50},
			toCosts:    map[string]float64{"EC2": 100},
			wantTotal:  100,
			wantDiff:   -50,
			wantPct:    -33.33333333333333,
			wantItems:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compare(tt.fromCosts, tt.toCosts, fromPeriod, toPeriod)

			if result.ToTotal != tt.wantTotal {
				t.Errorf("ToTotal = %v, want %v", result.ToTotal, tt.wantTotal)
			}
			if result.TotalDiff != tt.wantDiff {
				t.Errorf("TotalDiff = %v, want %v", result.TotalDiff, tt.wantDiff)
			}
			if result.TotalPct != tt.wantPct {
				t.Errorf("TotalPct = %v, want %v", result.TotalPct, tt.wantPct)
			}
			if len(result.Items) != tt.wantItems {
				t.Errorf("Items count = %v, want %v", len(result.Items), tt.wantItems)
			}
		})
	}
}

func TestCompare_NewService(t *testing.T) {
	fromPeriod := Period{Start: time.Now(), End: time.Now().AddDate(0, 1, 0)}
	toPeriod := Period{Start: time.Now().AddDate(0, 1, 0), End: time.Now().AddDate(0, 2, 0)}

	fromCosts := map[string]float64{"EC2": 100}
	toCosts := map[string]float64{"EC2": 100, "Lambda": 50}

	result := Compare(fromCosts, toCosts, fromPeriod, toPeriod)

	var lambdaItem *Item
	for i := range result.Items {
		if result.Items[i].Name == "Lambda" {
			lambdaItem = &result.Items[i]
			break
		}
	}

	if lambdaItem == nil {
		t.Fatal("Lambda item not found")
	}
	if !lambdaItem.IsNew {
		t.Error("Lambda should be marked as new")
	}
	if lambdaItem.DiffPct != 100 {
		t.Errorf("New service DiffPct = %v, want 100", lambdaItem.DiffPct)
	}
}

func TestCompare_RemovedService(t *testing.T) {
	fromPeriod := Period{Start: time.Now(), End: time.Now().AddDate(0, 1, 0)}
	toPeriod := Period{Start: time.Now().AddDate(0, 1, 0), End: time.Now().AddDate(0, 2, 0)}

	fromCosts := map[string]float64{"EC2": 100, "Lambda": 50}
	toCosts := map[string]float64{"EC2": 100}

	result := Compare(fromCosts, toCosts, fromPeriod, toPeriod)

	var lambdaItem *Item
	for i := range result.Items {
		if result.Items[i].Name == "Lambda" {
			lambdaItem = &result.Items[i]
			break
		}
	}

	if lambdaItem == nil {
		t.Fatal("Lambda item not found")
	}
	if !lambdaItem.IsRemoved {
		t.Error("Lambda should be marked as removed")
	}
	if lambdaItem.DiffPct != -100 {
		t.Errorf("Removed service DiffPct = %v, want -100", lambdaItem.DiffPct)
	}
}

func TestCompare_SortedByAbsDiff(t *testing.T) {
	fromPeriod := Period{Start: time.Now(), End: time.Now().AddDate(0, 1, 0)}
	toPeriod := Period{Start: time.Now().AddDate(0, 1, 0), End: time.Now().AddDate(0, 2, 0)}

	fromCosts := map[string]float64{"A": 100, "B": 100, "C": 100}
	toCosts := map[string]float64{"A": 110, "B": 150, "C": 80} // diffs: 10, 50, -20

	result := Compare(fromCosts, toCosts, fromPeriod, toPeriod)

	// Should be sorted by absolute diff: B (50), C (20), A (10)
	if result.Items[0].Name != "B" {
		t.Errorf("First item should be B, got %s", result.Items[0].Name)
	}
	if result.Items[1].Name != "C" {
		t.Errorf("Second item should be C, got %s", result.Items[1].Name)
	}
	if result.Items[2].Name != "A" {
		t.Errorf("Third item should be A, got %s", result.Items[2].Name)
	}
}

func TestCompareSimple(t *testing.T) {
	fromCosts := map[string]float64{"EC2": 100, "S3": 50}
	toCosts := map[string]float64{"EC2": 150, "S3": 30}

	items := CompareSimple(fromCosts, toCosts)

	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	// Should be sorted by absolute diff: EC2 (50), S3 (20)
	if items[0].Name != "EC2" {
		t.Errorf("First item should be EC2, got %s", items[0].Name)
	}
	if items[0].Diff != 50 {
		t.Errorf("EC2 Diff = %v, want 50", items[0].Diff)
	}
}

func TestFilterByMinDiff(t *testing.T) {
	items := []Item{
		{Name: "A", Diff: 100},
		{Name: "B", Diff: -50},
		{Name: "C", Diff: 10},
		{Name: "D", Diff: -5},
	}

	filtered := FilterByMinDiff(items, 20)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 items, got %d", len(filtered))
	}

	names := make(map[string]bool)
	for _, item := range filtered {
		names[item.Name] = true
	}

	if !names["A"] || !names["B"] {
		t.Error("Expected A and B to be in filtered results")
	}
}

func TestFilterByMinCost(t *testing.T) {
	items := []Item{
		{Name: "A", FromCost: 100, ToCost: 50},
		{Name: "B", FromCost: 10, ToCost: 20},
		{Name: "C", FromCost: 5, ToCost: 80},
	}

	filtered := FilterByMinCost(items, 50)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 items, got %d", len(filtered))
	}

	names := make(map[string]bool)
	for _, item := range filtered {
		names[item.Name] = true
	}

	if !names["A"] || !names["C"] {
		t.Error("Expected A and C to be in filtered results")
	}
}

func TestSortByToCost(t *testing.T) {
	items := []Item{
		{Name: "A", ToCost: 50},
		{Name: "B", ToCost: 100},
		{Name: "C", ToCost: 25},
	}

	SortByToCost(items)

	if items[0].Name != "B" || items[1].Name != "A" || items[2].Name != "C" {
		t.Errorf("Unexpected order: %s, %s, %s", items[0].Name, items[1].Name, items[2].Name)
	}
}

func TestSortByDiff(t *testing.T) {
	items := []Item{
		{Name: "A", Diff: 10},
		{Name: "B", Diff: -50},
		{Name: "C", Diff: 25},
	}

	SortByDiff(items)

	if items[0].Name != "B" || items[1].Name != "C" || items[2].Name != "A" {
		t.Errorf("Unexpected order: %s, %s, %s", items[0].Name, items[1].Name, items[2].Name)
	}
}

func TestSortByDiffPercent(t *testing.T) {
	items := []Item{
		{Name: "A", DiffPct: 10},
		{Name: "B", DiffPct: -50},
		{Name: "C", DiffPct: 25},
	}

	SortByDiffPercent(items)

	if items[0].Name != "B" || items[1].Name != "C" || items[2].Name != "A" {
		t.Errorf("Unexpected order: %s, %s, %s", items[0].Name, items[1].Name, items[2].Name)
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-0.5, 0.5},
	}

	for _, tt := range tests {
		if got := abs(tt.input); got != tt.want {
			t.Errorf("abs(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

