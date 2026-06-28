package scheduler

// ─── Daily ───────────────────────────────────────────────────────

type dailyHandler struct{}

func (dailyHandler) Code() string    { return "daily" }
func (dailyHandler) IsOneShot() bool { return false }

func (dailyHandler) BuildJob(data map[string]any) *JobDef {
	t := str(data, "time", "09:00")
	h, m := parseTime(t)
	return DurationJobDef(durationTo(h, m))
}

func init() { Register(dailyHandler{}) }
