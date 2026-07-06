package timeutil

import (
	"testing"
	"time"
)

func TestNow_IsUTC(t *testing.T) {
	if n := Now(); n.Location() != time.UTC {
		t.Errorf("Now().Location() = %s, want UTC", n.Location())
	}
}

func TestNow_IsRecent(t *testing.T) {
	n := Now()
	if d := time.Since(n); d < 0 || d > 5*time.Second {
		t.Errorf("Now() is %v old", d)
	}
}

func TestToday_IsMidnightUTC(t *testing.T) {
	td := Today()
	if td.Hour() != 0 || td.Minute() != 0 || td.Second() != 0 {
		t.Errorf("Today() = %v, want midnight", td)
	}
}

func TestToday_SameDayAsNow(t *testing.T) {
	td := Today()
	n := Now()
	if td.Year() != n.Year() || td.Month() != n.Month() || td.Day() != n.Day() {
		t.Errorf("Today()=%v, Now()=%v", td, n)
	}
}

func TestUTC_Converts(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	sh := time.Date(2026, 7, 6, 12, 0, 0, 0, loc)
	utc := UTC(sh)
	if utc.Location() != time.UTC {
		t.Error("not UTC")
	}
	if utc.Hour() != 4 {
		t.Errorf("12:00 CST => %d UTC, want 4", utc.Hour())
	}
}

func TestUTC_AlreadyUTC(t *testing.T) {
	orig := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	if !UTC(orig).Equal(orig) {
		t.Error("should not change UTC time")
	}
}
