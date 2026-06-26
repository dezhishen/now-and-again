package scheduler

import (
	"fmt"
	"time"
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

// durationTo computes the duration until next HH:MM.
func durationTo(h, m int) time.Duration {
	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
	if target.Before(now) {
		target = target.Add(24 * time.Hour)
	}
	return target.Sub(now)
}
