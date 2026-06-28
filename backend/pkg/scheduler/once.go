package scheduler

import (
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
)

// ─── Once (one-shot) ─────────────────────────────────────────────

type onceHandler struct{}

func (onceHandler) Code() string { return "once" }

func (onceHandler) IsOneShot() bool { return true }

func (onceHandler) BuildJob(data map[string]any) *JobDef {
	dateStr := str(data, "date", "")
	timeStr := str(data, "time", "00:00")

	// Parse "YYYY-MM-DD" + "HH:MM" as UTC (dates from API are timezone-agnostic)
	t, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+timeStr, time.UTC)
	if err != nil {
		// Fallback: fire immediately if parse fails
		t = timeutil.Now().Add(time.Minute)
	}
	// If the time is in the past, fire in 10 seconds (so it triggers almost immediately)
	if t.Before(timeutil.Now()) {
		t = timeutil.Now().Add(10 * time.Second)
	}
	return OneTimeJobDef(t)
}

func init() { Register(onceHandler{}) }
