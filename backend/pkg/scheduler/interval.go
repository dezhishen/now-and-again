package scheduler

import "time"

// ─── Interval ────────────────────────────────────────────────────

type intervalHandler struct{}

func (intervalHandler) Code() string { return "interval" }

func (intervalHandler) BuildJob(data map[string]any) *JobDef {
	days := 1
	if d, ok := data["days"].(float64); ok && int(d) > 0 {
		days = int(d)
	}
	return DurationJobDef(time.Duration(days) * 24 * time.Hour)
}

func init() { Register(intervalHandler{}) }
