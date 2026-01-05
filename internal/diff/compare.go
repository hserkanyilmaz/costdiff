package diff

import (
	"sort"
)

// Compare calculates the difference between two cost periods
func Compare(fromCosts, toCosts map[string]float64, fromPeriod, toPeriod Period) *Result {
	result := &Result{
		FromPeriod: fromPeriod,
		ToPeriod:   toPeriod,
		Items:      make([]Item, 0),
	}

	// Track all unique names
	allNames := make(map[string]bool)
	for name := range fromCosts {
		allNames[name] = true
	}
	for name := range toCosts {
		allNames[name] = true
	}

	// Build items
	for name := range allNames {
		fromCost := fromCosts[name]
		toCost := toCosts[name]

		item := Item{
			Name:     name,
			FromCost: fromCost,
			ToCost:   toCost,
			Diff:     toCost - fromCost,
		}

		// Calculate percentage change
		if fromCost == 0 && toCost > 0 {
			item.IsNew = true
			item.DiffPct = 100 // Treat new costs as 100% increase
		} else if fromCost > 0 && toCost == 0 {
			item.IsRemoved = true
			item.DiffPct = -100 // Treat removed costs as 100% decrease
		} else if fromCost > 0 {
			item.DiffPct = ((toCost - fromCost) / fromCost) * 100
		}

		result.Items = append(result.Items, item)
		result.FromTotal += fromCost
		result.ToTotal += toCost
	}

	// Calculate total diff
	result.TotalDiff = result.ToTotal - result.FromTotal
	if result.FromTotal > 0 {
		result.TotalPct = ((result.ToTotal - result.FromTotal) / result.FromTotal) * 100
	}

	// Sort by absolute diff (largest changes first)
	sort.Slice(result.Items, func(i, j int) bool {
		return abs(result.Items[i].Diff) > abs(result.Items[j].Diff)
	})

	return result
}

// CompareSimple compares two cost maps and returns items sorted by diff
func CompareSimple(fromCosts, toCosts map[string]float64) []Item {
	items := make([]Item, 0)

	// Track all unique names
	allNames := make(map[string]bool)
	for name := range fromCosts {
		allNames[name] = true
	}
	for name := range toCosts {
		allNames[name] = true
	}

	// Build items
	for name := range allNames {
		fromCost := fromCosts[name]
		toCost := toCosts[name]

		item := Item{
			Name:     name,
			FromCost: fromCost,
			ToCost:   toCost,
			Diff:     toCost - fromCost,
		}

		// Calculate percentage change
		if fromCost == 0 && toCost > 0 {
			item.IsNew = true
			item.DiffPct = 100
		} else if fromCost > 0 && toCost == 0 {
			item.IsRemoved = true
			item.DiffPct = -100
		} else if fromCost > 0 {
			item.DiffPct = ((toCost - fromCost) / fromCost) * 100
		}

		items = append(items, item)
	}

	// Sort by absolute diff (largest changes first)
	sort.Slice(items, func(i, j int) bool {
		return abs(items[i].Diff) > abs(items[j].Diff)
	})

	return items
}

// SortByToCost sorts items by current period cost descending
func SortByToCost(items []Item) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].ToCost > items[j].ToCost
	})
}

// SortByDiff sorts items by absolute diff descending
func SortByDiff(items []Item) {
	sort.Slice(items, func(i, j int) bool {
		return abs(items[i].Diff) > abs(items[j].Diff)
	})
}

// SortByDiffPercent sorts items by absolute diff percentage descending
func SortByDiffPercent(items []Item) {
	sort.Slice(items, func(i, j int) bool {
		return abs(items[i].DiffPct) > abs(items[j].DiffPct)
	})
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// FilterByMinDiff returns items with absolute diff >= minDiff
func FilterByMinDiff(items []Item, minDiff float64) []Item {
	filtered := make([]Item, 0)
	for _, item := range items {
		if abs(item.Diff) >= minDiff {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// FilterByMinCost returns items where either from or to cost >= minCost
func FilterByMinCost(items []Item, minCost float64) []Item {
	filtered := make([]Item, 0)
	for _, item := range items {
		if item.FromCost >= minCost || item.ToCost >= minCost {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

