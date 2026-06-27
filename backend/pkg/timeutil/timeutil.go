// Package timeutil provides centralized, UTC-enforced time functions.
//
// All internal timestamp handling MUST use this package instead of raw time.Now().
// This guarantees consistent UTC storage and computation regardless of deployment
// environment (Docker, bare metal, different timezone servers).
//
// For display purposes, use FormatInLocation() with the user's preferred timezone.
package timeutil

import "time"

// Now returns the current time in UTC.
// Use this everywhere instead of time.Now().
func Now() time.Time {
	return time.Now().UTC()
}

// Today returns the start of today (00:00:00 UTC).
func Today() time.Time {
	now := Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// UTC ensures the given time is in UTC. Returns t if already UTC, otherwise
// converts and returns a new time in UTC.
func UTC(t time.Time) time.Time {
	return t.UTC()
}

// FormatInLocation formats a UTC time for display in the given IANA timezone
// (e.g. "Asia/Shanghai", "America/New_York").
func FormatInLocation(t time.Time, tz string, layout string) string {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	return t.In(loc).Format(layout)
}

// ParseInLocation parses a time string as if it were in the given timezone,
// then converts and returns it in UTC. This is the correct way to interpret
// user input like "2026-06-27 09:00" which means 9 AM in their local time.
func ParseInLocation(layout, value, tz string) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	t, err := time.ParseInLocation(layout, value, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// LoadLocation safely loads an IANA timezone, falling back to UTC.
func LoadLocation(tz string) *time.Location {
	if tz == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}
