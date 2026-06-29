package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
)

// ─── Calendar types ───────────────────────────────────────────────

type CalendarDay struct {
	Date           string          `json:"date"`
	Weekday        int             `json:"weekday"`
	IsCurrentMonth bool            `json:"isCurrentMonth"`
	Events         []CalendarEvent `json:"events"`
}

type CalendarEvent struct {
	TaskID       string `json:"task_id"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Time         string `json:"time"`
	ScheduleType string `json:"schedule_type"`
	GroupName    string `json:"group_name,omitempty"`
}

// ─── CalendarService ──────────────────────────────────────────────

type CalendarService struct {
	repo *repository.TaskRepo
}

func NewCalendarService(repo *repository.TaskRepo) *CalendarService {
	return &CalendarService{repo: repo}
}

func (s *CalendarService) GetCalendar(ctx context.Context, year, month int, groupID string) (any, error) {
	familyID, _ := ctx.Value("family_id").(string)

	tasks, err := s.repo.ListTasksByFamily(familyID)
	if err != nil {
		return nil, err
	}

	loc := time.UTC
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	lastDay := firstDay.AddDate(0, 1, -1)

	dayEvents := make(map[int][]CalendarEvent)

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		if groupID != "" && task.GroupID != "" && task.GroupID != groupID {
			continue
		}

		eventTime := parseEventTime(task.ScheduleData)
		days := expandSchedule(task.ScheduleType, task.ScheduleData, task.CreatedAt, year, month)
		for _, d := range days {
			dayEvents[d] = append(dayEvents[d], CalendarEvent{
				TaskID:       task.ID,
				Name:         task.Name,
				Kind:         task.Kind,
				Time:         eventTime,
				ScheduleType: task.ScheduleType,
				GroupName:    task.Group.Name,
			})
		}
	}

	var result []CalendarDay
	startWeekday := int(firstDay.Weekday())

	prevMonth := firstDay.AddDate(0, 0, -1)
	prevLast := prevMonth.Day()
	for i := startWeekday - 1; i >= 0; i-- {
		d := prevLast - i
		date := time.Date(year, time.Month(month)-1, d, 0, 0, 0, 0, loc)
		result = append(result, CalendarDay{
			Date:           date.Format("2006-01-02"),
			Weekday:        int(date.Weekday()),
			IsCurrentMonth: false,
			Events:         []CalendarEvent{},
		})
	}

	for d := 1; d <= lastDay.Day(); d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, loc)
		events := dayEvents[d]
		if events == nil {
			events = []CalendarEvent{}
		}
		result = append(result, CalendarDay{
			Date:           date.Format("2006-01-02"),
			Weekday:        int(date.Weekday()),
			IsCurrentMonth: true,
			Events:         events,
		})
	}

	remaining := 42 - len(result)
	for d := 1; d <= remaining; d++ {
		date := time.Date(year, time.Month(month)+1, d, 0, 0, 0, 0, loc)
		result = append(result, CalendarDay{
			Date:           date.Format("2006-01-02"),
			Weekday:        int(date.Weekday()),
			IsCurrentMonth: false,
			Events:         []CalendarEvent{},
		})
	}

	return result, nil
}

// ─── Helpers ──────────────────────────────────────────────────────

func parseEventTime(scheduleData string) string {
	var data map[string]any
	json.Unmarshal([]byte(scheduleData), &data)
	if t, ok := data["time"].(string); ok && t != "" {
		return t
	}
	return "09:00"
}

func expandSchedule(scheduleType, scheduleData string, createdAt time.Time, year, month int) []int {
	var data map[string]any
	json.Unmarshal([]byte(scheduleData), &data)

	switch scheduleType {
	case "daily":
		return allDays(month, year)
	case "weekly":
		return weeklyDays(data, createdAt, year, month)
	case "monthly":
		return monthlyDays(data, year, month)
	case "interval":
		return intervalDays(data, createdAt, year, month)
	case "once":
		return onceDay(data, year, month)
	default:
		return nil
	}
}

func allDays(month, year int) []int {
	lastDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1).Day()
	days := make([]int, lastDay)
	for i := range days {
		days[i] = i + 1
	}
	return days
}

func weeklyDays(data map[string]any, createdAt time.Time, year, month int) []int {
	raw, _ := data["days"].([]any)
	weekdays := make(map[int]bool)
	for _, v := range raw {
		if n, ok := v.(float64); ok {
			wd := int(n)
			if wd == 7 {
				wd = 0
			}
			weekdays[wd] = true
		}
	}
	if len(weekdays) == 0 {
		weekdays[int(createdAt.Weekday())] = true
	}

	var days []int
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		if weekdays[int(d.Weekday())] {
			days = append(days, d.Day())
		}
	}
	return days
}

func monthlyDays(data map[string]any, year, month int) []int {
	raw, _ := data["days"].([]any)
	daySet := make(map[int]bool)
	for _, v := range raw {
		if n, ok := v.(float64); ok {
			daySet[int(n)] = true
		}
	}
	if len(daySet) == 0 {
		daySet[1] = true
	}

	lastDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1).Day()
	var days []int
	for d := 1; d <= lastDay; d++ {
		if daySet[d] {
			days = append(days, d)
		}
	}
	return days
}

func intervalDays(data map[string]any, createdAt time.Time, year, month int) []int {
	interval := 1
	if d, ok := data["days"].(float64); ok && int(d) > 0 {
		interval = int(d)
	}

	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)

	var days []int
	current := createdAt
	for !current.After(last) {
		if !current.Before(first) {
			days = append(days, current.Day())
		}
		current = current.AddDate(0, 0, interval)
	}
	return days
}

func onceDay(data map[string]any, year, month int) []int {
	ds, _ := data["date"].(string)
	if ds == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", ds, time.UTC)
	if err != nil {
		return nil
	}
	if t.Year() == year && int(t.Month()) == month {
		return []int{t.Day()}
	}
	return nil
}
