package scheduler

import (
	"fmt"
	"strings"
)

// ─── Monthly ─────────────────────────────────────────────────────

type monthlyHandler struct{}

func (monthlyHandler) Code() string { return "monthly" }

func (monthlyHandler) BuildJob(data map[string]any) *JobDef {
	t := str(data, "time", "09:00")
	h, m := parseTime(t)
	days := ints(data, "days")
	if len(days) == 0 {
		days = []int{1}
	}
	dayStrs := make([]string, len(days))
	for i, d := range days {
		dayStrs[i] = fmt.Sprintf("%d", d)
	}
	expr := fmt.Sprintf("%d %d %s * *", m, h, strings.Join(dayStrs, ","))
	return CronJobDef(expr)
}

func init() { Register(monthlyHandler{}) }
