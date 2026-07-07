package scheduler

import (
	"encoding/json"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/dezhishen/now-and-again/backend/pkg/scheduler/engine"
	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
)

// ─── Once (one-shot) ─────────────────────────────────────────────

type onceHandler struct{}

func (onceHandler) Code() string { return "once" }

func (h onceHandler) Schedule(t TaskInfo) error {
	var data map[string]any
	json.Unmarshal([]byte(t.ScheduleData), &data)
	def := h.buildJob(data)

	// One-shot task function: auto-remove after successful fire.
	taskFn := gocron.NewTask(func() {
		now := timeutil.Now()
		log("triggered", t.TaskID, "")
		if err := t.OnFire(t.TaskID, now); err != nil {
			log("error", t.TaskID, err.Error())
			return
		}
		engine.Get().RemoveJob(t.TaskID)
		mu.Lock()
		delete(scheduled, t.TaskID)
		mu.Unlock()
		if t.OnDone != nil {
			t.OnDone(t.TaskID)
		}
	})

	return engine.Get().AddJob(t.TaskID, def, taskFn)
}

func (onceHandler) Unschedule(taskID string) {
	engine.Get().RemoveJob(taskID)
}

func (onceHandler) OnManualComplete(taskID string, onDone func(taskID string)) {
	engine.Get().RemoveJob(taskID)
	mu.Lock()
	delete(scheduled, taskID)
	mu.Unlock()
	log("completed", taskID, "")
	if onDone != nil {
		onDone(taskID)
	}
}

func (onceHandler) buildJob(data map[string]any) gocron.JobDefinition {
	dateStr := str(data, "date", "")
	timeStr := str(data, "time", "00:00")

	t, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+timeStr, time.UTC)
	if err != nil {
		t = timeutil.Now().Add(time.Minute)
	}
	// 已过期的 one-shot 任务不再重调度（避免重启时重复生成待办）
	// 调用方（registerToScheduler）应在注册前自行过滤此类任务
	if !t.After(timeutil.Now()) {
		t = timeutil.Now().Add(100 * 365 * 24 * time.Hour) // 推到遥远的未来，永不触发
	}
	return engine.OneTimeJobDef(t)
}

func init() { Register(onceHandler{}) }
