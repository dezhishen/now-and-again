package scheduler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/pkg/scheduler/engine"
)

// ─── Yearly ──────────────────────────────────────────────────────

type yearlyHandler struct{}

func (yearlyHandler) Code() string { return "yearly" }

func (h yearlyHandler) Schedule(t TaskInfo) error {
	var data map[string]any
	json.Unmarshal([]byte(t.ScheduleData), &data)
	def := h.buildJob(data)
	taskFn := gocron.NewTask(defaultTaskFn(t))
	return engine.Get().AddJob(t.TaskID, def, taskFn)
}

func (yearlyHandler) Unschedule(taskID string) {
	engine.Get().RemoveJob(taskID)
}

func (yearlyHandler) OnManualComplete(string, func(string)) {}

func (yearlyHandler) buildJob(data map[string]any) gocron.JobDefinition {
	t := str(data, "time", "09:00")
	h, m := parseTime(t)
	months := ints(data, "days") // reuse "days" format like monthly
	if len(months) == 0 {
		months = []int{1}
	}
	// Yearly: month(s) + day. Default day = 1.
	day := 1
	if v, ok := data["day"]; ok {
		switch n := v.(type) {
		case float64:
			day = int(n)
		case int:
			day = n
		}
	}
	monthStrs := make([]string, len(months))
	for i, mo := range months {
		monthStrs[i] = fmt.Sprintf("%d", mo)
	}
	// gocron cron: min hour day month *
	expr := fmt.Sprintf("%d %d %d %s *", m, h, day, strings.Join(monthStrs, ","))
	return engine.CronJobDef(expr)
}

func init() { Register(yearlyHandler{}) }
