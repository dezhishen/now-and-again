// Package sdk — timezone-aware conversions.
//
// All times stored and served by the backend are in UTC.
// When the CLI (or any SDK consumer) accepts time input from a user,
// the input is in the user's local timezone and MUST be converted to UTC
// before sending to the backend.
//
// Conversely, when displaying times from the backend, they MUST be converted
// to the user's local timezone.
//
// This mirrors the frontend's timezone.ts logic in Go.

package sdk

import (
	"fmt"
	"regexp"
	"time"
)

// ─── Timezone helpers on NA ──────────────────────────────────────

// GetTimezone returns the configured timezone location.
// Defaults to time.Local if none was set.
func (na *NA) GetTimezone() *time.Location {
	na.mu.RLock()
	defer na.mu.RUnlock()
	if na.timezone != nil {
		return na.timezone
	}
	return time.Local
}

// SetTimezone sets the timezone to use for all conversions.
func (na *NA) SetTimezone(loc *time.Location) {
	na.mu.Lock()
	defer na.mu.Unlock()
	na.timezone = loc
}

// ─── Formatting ──────────────────────────────────────────────────

// FormatTime formats a UTC time for display in the configured timezone.
func (na *NA) FormatTime(t time.Time, layout string) string {
	return t.In(na.GetTimezone()).Format(layout)
}

// FormatTimeIn formats a UTC time in the given location.
func FormatTimeIn(t time.Time, loc *time.Location, layout string) string {
	return t.In(loc).Format(layout)
}

// ─── HH:MM conversions ──────────────────────────────────────────

// localTimeToUTC converts a local-wall-clock "HH:MM" string to UTC "HH:MM".
// Uses today's date as the anchor.
func localTimeToUTC(localTime string, loc *time.Location) (string, error) {
	h, m, err := parseHM(localTime)
	if err != nil {
		return "", err
	}
	now := time.Now().In(loc)
	local := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, loc)
	utc := local.UTC()
	return fmt.Sprintf("%02d:%02d", utc.Hour(), utc.Minute()), nil
}

// utcTimeToLocal converts a UTC "HH:MM" string to local "HH:MM".
func utcTimeToLocal(utcTime string, loc *time.Location) (string, error) {
	h, m, err := parseHM(utcTime)
	if err != nil {
		return "", err
	}
	now := time.Now().In(loc)
	utc := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.UTC)
	local := utc.In(loc)
	return fmt.Sprintf("%02d:%02d", local.Hour(), local.Minute()), nil
}

// ─── Date+Time conversions ──────────────────────────────────────

// localDateTimeToUTC converts local "YYYY-MM-DD" + "HH:MM" to UTC {date, time}.
func localDateTimeToUTC(localDate, localTime string, loc *time.Location) (date string, tim string, err error) {
	h, m, e := parseHM(localTime)
	if e != nil {
		return "", "", e
	}
	d, e := time.ParseInLocation("2006-01-02", localDate, loc)
	if e != nil {
		return "", "", fmt.Errorf("invalid date %q: %w", localDate, e)
	}
	local := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, loc)
	utc := local.UTC()
	date = utc.Format("2006-01-02")
	tim = fmt.Sprintf("%02d:%02d", utc.Hour(), utc.Minute())
	return date, tim, nil
}

// utcDateTimeToLocal converts UTC "YYYY-MM-DD" + "HH:MM" to local {date, time}.
func utcDateTimeToLocal(utcDate, utcTime string, loc *time.Location) (date string, tim string, err error) {
	h, m, e := parseHM(utcTime)
	if e != nil {
		return "", "", e
	}
	d, e2 := time.ParseInLocation("2006-01-02", utcDate, time.UTC)
	if e2 != nil {
		return "", "", fmt.Errorf("invalid utc date %q: %w", utcDate, e2)
	}
	utc := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, time.UTC)
	local := utc.In(loc)
	date = local.Format("2006-01-02")
	tim = fmt.Sprintf("%02d:%02d", local.Hour(), local.Minute())
	return date, tim, nil
}

// ─── Schedule data deep conversion ──────────────────────────────

// scheduleDataToUTC converts known time fields in schedule_data
// from local → UTC. The value must be a map[string]interface{}.
func scheduleDataToUTC(sd any, loc *time.Location) any {
	m, ok := sd.(map[string]interface{})
	if !ok {
		return sd
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	// Convert standalone time field
	if timeStr, ok := out["time"].(string); ok && isValidHM(timeStr) {
		if utcTime, err := localTimeToUTC(timeStr, loc); err == nil {
			out["time"] = utcTime
		}
	}
	// Convert date+time for one-shot tasks
	if dateStr, ok := out["date"].(string); ok && isValidDate(dateStr) {
		if timeStr, ok := out["time"].(string); ok && isValidHM(timeStr) {
			if dt, tm, err := localDateTimeToUTC(dateStr, timeStr, loc); err == nil {
				out["date"] = dt
				out["time"] = tm
			}
		}
	}
	return out
}

// scheduleDataToLocal converts known time fields in schedule_data
// from UTC → local. The value must be a map[string]interface{}.
func scheduleDataToLocal(sd any, loc *time.Location) any {
	m, ok := sd.(map[string]interface{})
	if !ok {
		return sd
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	// Convert standalone time field
	if timeStr, ok := out["time"].(string); ok && isValidHM(timeStr) {
		if localTime, err := utcTimeToLocal(timeStr, loc); err == nil {
			out["time"] = localTime
		}
	}
	// Convert date+time for one-shot tasks
	if dateStr, ok := out["date"].(string); ok && isValidDate(dateStr) {
		if timeStr, ok := out["time"].(string); ok && isValidHM(timeStr) {
			if dt, tm, err := utcDateTimeToLocal(dateStr, timeStr, loc); err == nil {
				out["date"] = dt
				out["time"] = tm
			}
		}
	}
	return out
}

// ─── Parsing helpers ─────────────────────────────────────────────

func parseHM(s string) (h, m int, err error) {
	_, err = fmt.Sscanf(s, "%d:%d", &h, &m)
	if err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("invalid HH:MM time %q", s)
	}
	return h, m, nil
}

var (
	hmRe   = regexp.MustCompile(`^\d{2}:\d{2}$`)
	dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func isValidHM(s string) bool  { return hmRe.MatchString(s) }
func isValidDate(s string) bool { return dateRe.MatchString(s) }
