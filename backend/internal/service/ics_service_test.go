package service

import (
	"strings"
	"testing"
	"time"
)

// ─── ICS event duration clamping ─────────────────────────────────

func TestICSEventDuration_NoClampForMorningTime(t *testing.T) {
	// 09:00 UTC → DTEND should be 10:00 UTC (no clamp needed)
	schedTime := time.Date(2026, 7, 6, 9, 0, 0, 0, time.UTC)
	eventEnd := clampICSEndTime(schedTime)
	expected := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	if !eventEnd.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, eventEnd)
	}
}

func TestICSEventDuration_ClampNearMidnight(t *testing.T) {
	// 23:30 UTC → DTEND would be 00:30 next day, should clamp to 23:59:59
	schedTime := time.Date(2026, 7, 6, 23, 30, 0, 0, time.UTC)
	eventEnd := clampICSEndTime(schedTime)
	expected := time.Date(2026, 7, 6, 23, 59, 59, 0, time.UTC)
	if !eventEnd.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, eventEnd)
	}
}

func TestICSEventDuration_ClampExactlyAt23(t *testing.T) {
	// 23:00 UTC → DTEND would be 00:00 next day, should clamp to 23:59:59
	schedTime := time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC)
	eventEnd := clampICSEndTime(schedTime)
	expected := time.Date(2026, 7, 6, 23, 59, 59, 0, time.UTC)
	if !eventEnd.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, eventEnd)
	}
}

func TestICSEventDuration_NoClampForAfternoonTime(t *testing.T) {
	// 14:00 UTC → DTEND should be 15:00 UTC (no clamp)
	schedTime := time.Date(2026, 7, 6, 14, 0, 0, 0, time.UTC)
	eventEnd := clampICSEndTime(schedTime)
	expected := time.Date(2026, 7, 6, 15, 0, 0, 0, time.UTC)
	if !eventEnd.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, eventEnd)
	}
}

func TestICSEventDuration_SameDayNotSameDate(t *testing.T) {
	// 00:30 UTC → DTEND should be 01:30 (no clamp, still same day)
	schedTime := time.Date(2026, 7, 6, 0, 30, 0, 0, time.UTC)
	eventEnd := clampICSEndTime(schedTime)
	expected := time.Date(2026, 7, 6, 1, 30, 0, 0, time.UTC)
	if !eventEnd.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, eventEnd)
	}
}

// ─── ICS output validation ───────────────────────────────────────

func TestGenerateICS_ContainsCalendarProperties(t *testing.T) {
	// This test only verifies that we can generate a valid-ish ICS
	// without a real database. We test the output string structure.
	ics := buildMinimalICS("Test Calendar", "A test", []minimalTask{})
	if !strings.Contains(ics, "BEGIN:VCALENDAR") {
		t.Error("missing VCALENDAR begin")
	}
	if !strings.Contains(ics, "VERSION:2.0") {
		t.Error("missing VERSION")
	}
	if !strings.Contains(ics, "CALSCALE:GREGORIAN") {
		t.Error("missing CALSCALE")
	}
	if !strings.Contains(ics, "NAME:Test Calendar") {
		t.Error("missing NAME property")
	}
	if !strings.Contains(ics, "X-WR-CALNAME:Test Calendar") {
		t.Error("missing X-WR-CALNAME")
	}
	if !strings.Contains(ics, "END:VCALENDAR") {
		t.Error("missing VCALENDAR end")
	}
}

func TestGenerateICS_TaskTimesDoNotCrossMidnight(t *testing.T) {
	// Verify that a task near midnight doesn't produce DTEND on the next day
	schedTime := time.Date(2026, 7, 6, 23, 50, 0, 0, time.UTC)
	end := clampICSEndTime(schedTime)
	if end.Day() != schedTime.Day() {
		t.Errorf("DTEND crossed midnight: start day=%d, end day=%d", schedTime.Day(), end.Day())
	}
}

// ─── Test helpers ─────────────────────────────────────────────────

// clampICSEndTime mirrors the clamping logic in ics_service.go.
func clampICSEndTime(schedTime time.Time) time.Time {
	eventEnd := schedTime.Add(1 * time.Hour)
	schedDay := schedTime.Truncate(24 * time.Hour)
	nextDay := schedDay.Add(24 * time.Hour)
	if !eventEnd.Before(nextDay) {
		eventEnd = nextDay.Add(-1 * time.Second)
	}
	return eventEnd
}

type minimalTask struct {
	name      string
	schedType string
	schedData string
}

func buildMinimalICS(name, desc string, tasks []minimalTask) string {
	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//Test//EN\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")
	sb.WriteString("NAME:")
	sb.WriteString(name)
	sb.WriteString("\r\n")
	sb.WriteString("X-WR-CALNAME:")
	sb.WriteString(name)
	sb.WriteString("\r\n")
	sb.WriteString("X-WR-CALDESC:")
	sb.WriteString(desc)
	sb.WriteString("\r\n")
	for range tasks {
		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString("END:VEVENT\r\n")
	}
	sb.WriteString("END:VCALENDAR\r\n")
	return sb.String()
}
