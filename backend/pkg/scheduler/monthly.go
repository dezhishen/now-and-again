package scheduler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/pkg/scheduler/engine"
)

// ─── Monthly ─────────────────────────────────────────────────────

type monthlyHandler struct{}

func (monthlyHandler) Code() string { return "monthly" }

func (h monthlyHandler) Schedule(t TaskInfo) error {
	var data map[string]any
	json.Unmarshal([]byte(t.ScheduleData), &data)
	def := h.buildJob(data)
	taskFn := gocron.NewTask(defaultTaskFn(t))
	return engine.Get().AddJob(t.TaskID, def, taskFn)
}

func (monthlyHandler) Unschedule(taskID string) {
	engine.Get().RemoveJob(taskID)
}

func (monthlyHandler) OnManualComplete(string, func(string)) {}

func (monthlyHandler) buildJob(data map[string]any) gocron.JobDefinition {
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
	return engine.CronJobDef(expr)
}

func init() { Register(monthlyHandler{}) }
