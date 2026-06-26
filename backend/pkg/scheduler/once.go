package scheduler

import "time"

// ─── Once (one-shot) ─────────────────────────────────────────────

type onceHandler struct{}

func (onceHandler) Code() string { return "once" }

func (onceHandler) BuildJob(data map[string]any) *JobDef {
	dateStr := str(data, "date", "")
	timeStr := str(data, "time", "00:00")

	// Parse "YYYY-MM-DD" + "HH:MM"
	t, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+timeStr, time.Local)
	if err != nil {
		// Fallback: fire immediately if parse fails
		t = time.Now().Add(time.Minute)
	}
	// If the time is in the past, fire in 10 seconds (so it triggers almost immediately)
	if t.Before(time.Now()) {
		t = time.Now().Add(10 * time.Second)
	}
	return OneTimeJobDef(t)
}

func init() { Register(onceHandler{}) }
