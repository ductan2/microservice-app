package types

import "time"

// TimeRange represents a time period for filtering data
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}