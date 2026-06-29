package scheduler

import (
	"fmt"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
)

// str safely gets a string from a map.
func str(data map[string]any, key, fallback string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return fallback
}

// ints safely gets []int from a map (stored as []any of float64).
func ints(data map[string]any, key string) []int {
	arr, ok := data[key].([]any)
	if !ok {
		return nil
	}
	result := make([]int, len(arr))
	for i, item := range arr {
		if n, ok := item.(float64); ok {
			result[i] = int(n)
		}
	}
	return result
}

// parseTime parses "HH:MM".
func parseTime(t string) (h, m int) {
	fmt.Sscanf(t, "%d:%d", &h, &m)
	return
}

// durationTo computes the duration until next HH:MM (in UTC).
func durationTo(h, m int) time.Duration {
	now := timeutil.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.UTC)
	if !target.After(now) {
		target = target.Add(24 * time.Hour)
	}
	return target.Sub(now)
}

// DurationToInLocation computes the duration until next HH:MM in the given timezone.
func DurationToInLocation(h, m int, tz string) time.Duration {
	loc := timeutil.LoadLocation(tz)
	now := timeutil.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, loc)
	if target.Before(now) {
		target = target.Add(24 * time.Hour)
	}
	return target.Sub(now)
}
