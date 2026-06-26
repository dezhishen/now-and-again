package scheduler

import (
	"fmt"
	"strings"
)

// ─── Weekly ──────────────────────────────────────────────────────

type weeklyHandler struct{}

func (weeklyHandler) Code() string { return "weekly" }

func (weeklyHandler) BuildJob(data map[string]any) *JobDef {
	t := str(data, "time", "09:00")
	h, m := parseTime(t)
	days := ints(data, "days")
	if len(days) == 0 {
		days = []int{1}
	}
	// Build cron: minute hour * * day-of-week
	dayStrs := make([]string, len(days))
	for i, d := range days {
		// gocron cron: Sunday=0, we use Monday=1..Sunday=7
		if d == 7 {
			d = 0
		}
		dayStrs[i] = fmt.Sprintf("%d", d)
	}
	expr := fmt.Sprintf("%d %d * * %s", m, h, strings.Join(dayStrs, ","))
	return CronJobDef(expr)
}

func init() { Register(weeklyHandler{}) }
