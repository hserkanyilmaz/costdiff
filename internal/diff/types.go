package diff

import (
	"time"
)

// Period represents a time period for cost comparison
type Period struct {
	Start time.Time
	End   time.Time
}

// Label returns a human-readable label for the period
func (p Period) Label() string {
	// If period is exactly one month
	if p.Start.Day() == 1 && p.End.Day() == 1 {
		nextMonth := p.Start.AddDate(0, 1, 0)
		if p.End.Equal(nextMonth) {
			return p.Start.Format("Jan 2006")
		}
	}

	// If period is a single day
	if p.End.Sub(p.Start) == 24*time.Hour {
		return p.Start.Format("Jan 2, 2006")
	}

	// Generic range
	return p.Start.Format("Jan 2") + " - " + p.End.AddDate(0, 0, -1).Format("Jan 2, 2006")
}

// Item represents a single cost item with comparison data
type Item struct {
	Name       string  `json:"name"`
	FromCost   float64 `json:"from_cost"`
	ToCost     float64 `json:"to_cost"`
	Diff       float64 `json:"diff"`
	DiffPct    float64 `json:"diff_percent"`
	IsNew      bool    `json:"is_new,omitempty"`
	IsRemoved  bool    `json:"is_removed,omitempty"`
}

// Result represents the complete comparison result
type Result struct {
	FromPeriod Period  `json:"from_period"`
	ToPeriod   Period  `json:"to_period"`
	FromTotal  float64 `json:"from_total"`
	ToTotal    float64 `json:"to_total"`
	TotalDiff  float64 `json:"total_diff"`
	TotalPct   float64 `json:"total_diff_percent"`
	Items      []Item  `json:"items"`
}

// TopItem represents a single cost item for the top command
type TopItem struct {
	Name    string  `json:"name"`
	Cost    float64 `json:"cost"`
	Percent float64 `json:"percent"`
}

// TopResult represents the result of the top command
type TopResult struct {
	Period Period    `json:"period"`
	Total  float64   `json:"total"`
	Items  []TopItem `json:"items"`
}

// DayItem represents a single day's cost
type DayItem struct {
	Date          time.Time `json:"date"`
	Cost          float64   `json:"cost"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
}

// WatchResult represents the result of the watch command
type WatchResult struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Total     float64   `json:"total"`
	Average   float64   `json:"average"`
	Days      []DayItem `json:"days"`
}

// PeriodJSON is a JSON-friendly representation of Period
type PeriodJSON struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Label string `json:"label"`
}

// ToJSON converts Period to PeriodJSON
func (p Period) ToJSON() PeriodJSON {
	return PeriodJSON{
		Start: p.Start.Format("2006-01-02"),
		End:   p.End.Format("2006-01-02"),
		Label: p.Label(),
	}
}

// ResultJSON is a JSON-friendly representation of Result
type ResultJSON struct {
	FromPeriod PeriodJSON `json:"from_period"`
	ToPeriod   PeriodJSON `json:"to_period"`
	FromTotal  float64    `json:"from_total"`
	ToTotal    float64    `json:"to_total"`
	TotalDiff  float64    `json:"total_diff"`
	TotalPct   float64    `json:"total_diff_percent"`
	Items      []Item     `json:"items"`
}

// ToJSON converts Result to ResultJSON
func (r *Result) ToJSON() ResultJSON {
	return ResultJSON{
		FromPeriod: r.FromPeriod.ToJSON(),
		ToPeriod:   r.ToPeriod.ToJSON(),
		FromTotal:  r.FromTotal,
		ToTotal:    r.ToTotal,
		TotalDiff:  r.TotalDiff,
		TotalPct:   r.TotalPct,
		Items:      r.Items,
	}
}

// TopResultJSON is a JSON-friendly representation of TopResult
type TopResultJSON struct {
	Period PeriodJSON `json:"period"`
	Total  float64    `json:"total"`
	Items  []TopItem  `json:"items"`
}

// ToJSON converts TopResult to TopResultJSON
func (r *TopResult) ToJSON() TopResultJSON {
	return TopResultJSON{
		Period: r.Period.ToJSON(),
		Total:  r.Total,
		Items:  r.Items,
	}
}

// WatchResultJSON is a JSON-friendly representation of WatchResult
type WatchResultJSON struct {
	StartDate string        `json:"start_date"`
	EndDate   string        `json:"end_date"`
	Total     float64       `json:"total"`
	Average   float64       `json:"average"`
	Days      []DayItemJSON `json:"days"`
}

// DayItemJSON is a JSON-friendly representation of DayItem
type DayItemJSON struct {
	Date          string  `json:"date"`
	Cost          float64 `json:"cost"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
}

// ToJSON converts WatchResult to WatchResultJSON
func (r *WatchResult) ToJSON() WatchResultJSON {
	days := make([]DayItemJSON, len(r.Days))
	for i, d := range r.Days {
		days[i] = DayItemJSON{
			Date:          d.Date.Format("2006-01-02"),
			Cost:          d.Cost,
			Change:        d.Change,
			ChangePercent: d.ChangePercent,
		}
	}

	return WatchResultJSON{
		StartDate: r.StartDate.Format("2006-01-02"),
		EndDate:   r.EndDate.Format("2006-01-02"),
		Total:     r.Total,
		Average:   r.Average,
		Days:      days,
	}
}

