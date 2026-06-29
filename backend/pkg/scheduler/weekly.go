package scheduler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/pkg/scheduler/engine"
)

// ─── Weekly ──────────────────────────────────────────────────────

type weeklyHandler struct{}

func (weeklyHandler) Code() string { return "weekly" }

func (h weeklyHandler) Schedule(t TaskInfo) error {
	var data map[string]any
	json.Unmarshal([]byte(t.ScheduleData), &data)
	def := h.buildJob(data)
	taskFn := gocron.NewTask(defaultTaskFn(t))
	return engine.Get().AddJob(t.TaskID, def, taskFn)
}

func (weeklyHandler) Unschedule(taskID string) {
	engine.Get().RemoveJob(taskID)
}

func (weeklyHandler) OnManualComplete(string, func(string)) {}

func (weeklyHandler) buildJob(data map[string]any) gocron.JobDefinition {
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
	return engine.CronJobDef(expr)
}

func init() { Register(weeklyHandler{}) }
