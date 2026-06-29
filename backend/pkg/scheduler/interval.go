package scheduler

import (
	"encoding/json"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/pkg/scheduler/engine"
)

// ─── Interval ────────────────────────────────────────────────────

type intervalHandler struct{}

func (intervalHandler) Code() string { return "interval" }

func (h intervalHandler) Schedule(t TaskInfo) error {
	var data map[string]any
	json.Unmarshal([]byte(t.ScheduleData), &data)
	def := h.buildJob(data)
	taskFn := gocron.NewTask(defaultTaskFn(t))
	return engine.Get().AddJob(t.TaskID, def, taskFn)
}

func (intervalHandler) Unschedule(taskID string) {
	engine.Get().RemoveJob(taskID)
}

func (intervalHandler) OnManualComplete(string, func(string)) {}

func (intervalHandler) buildJob(data map[string]any) gocron.JobDefinition {
	days := 1
	if d, ok := data["days"].(float64); ok && int(d) > 0 {
		days = int(d)
	}
	return engine.DurationJobDef(time.Duration(days) * 24 * time.Hour)
}

func init() { Register(intervalHandler{}) }
