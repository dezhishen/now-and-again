package scheduler

import (
	"encoding/json"
	"fmt"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/pkg/scheduler/engine"
)

// ─── Daily ───────────────────────────────────────────────────────

type dailyHandler struct{}

func (dailyHandler) Code() string { return "daily" }

func (h dailyHandler) Schedule(t TaskInfo) error {
	var data map[string]any
	json.Unmarshal([]byte(t.ScheduleData), &data)
	def := h.buildJob(data)
	taskFn := gocron.NewTask(defaultTaskFn(t))
	return engine.Get().AddJob(t.TaskID, def, taskFn)
}

func (dailyHandler) Unschedule(taskID string) {
	engine.Get().RemoveJob(taskID)
}

func (dailyHandler) OnManualComplete(string, func(string)) {}

func (dailyHandler) buildJob(data map[string]any) gocron.JobDefinition {
	t := str(data, "time", "09:00")
	h, m := parseTime(t)
	// Use CronJob so the job fires every day at the same HH:MM,
	// instead of DurationJob which drifts because its interval is
	// computed relative to "now" at registration time.
	expr := fmt.Sprintf("%d %d * * *", m, h)
	return engine.CronJobDef(expr)
}

func init() { Register(dailyHandler{}) }
