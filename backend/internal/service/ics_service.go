package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"
	"golang.org/x/crypto/bcrypt"
)

type IcsService struct {
	repo       *repository.IcsRepo
	taskRepo   *repository.TaskRepo
	apiKeyRepo *repository.ApiKeyRepo
	userRepo   *repository.UserRepo
}

func NewIcsService(repo *repository.IcsRepo, taskRepo *repository.TaskRepo, apiKeyRepo *repository.ApiKeyRepo, userRepo *repository.UserRepo) *IcsService {
	return &IcsService{repo: repo, taskRepo: taskRepo, apiKeyRepo: apiKeyRepo, userRepo: userRepo}
}

// ─── Feed CRUD ───────────────────────────────────────────────────

type CreateIcsFeedInput struct {
	Name          string
	Description   string
	FilterDays    int
	FilterGroupID string
	FilterType    string // "all" or "personal"
	AuthType      string // "api_key" or "basic"
	ApiKeyID      string
	AppPassword   string // only used when AuthType=basic
	FamilyID      string
	CreatedBy     string
}

func (s *IcsService) Create(input CreateIcsFeedInput) (*repository.IcsFeedModel, error) {
	if input.FilterDays <= 0 {
		input.FilterDays = 7
	}
	if input.FilterType == "" {
		input.FilterType = "all"
	}
	if input.AuthType == "" {
		input.AuthType = "api_key"
	}

	feed := &repository.IcsFeedModel{
		FamilyID:      input.FamilyID,
		Name:          input.Name,
		Description:   input.Description,
		FilterDays:    input.FilterDays,
		FilterGroupID: input.FilterGroupID,
		FilterType:    input.FilterType,
		AuthType:      input.AuthType,
		ApiKeyID:      input.ApiKeyID,
		Enabled:       true,
		AccessToken:   uuid.New().String(),
		CreatedBy:     input.CreatedBy,
	}

	// For basic auth, look up the user's account username and hash the password
	if input.AuthType == "basic" {
		username, err := s.getAccountUsername(input.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("find account: %w", err)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(input.AppPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		feed.AppUsername = username
		feed.AppPasswordHash = string(hash)
	}

	if err := s.repo.CreateFeed(feed); err != nil {
		return nil, err
	}
	return feed, nil
}

func (s *IcsService) List(familyID string) ([]repository.IcsFeedModel, error) {
	return s.repo.ListFeedsByFamily(familyID)
}

func (s *IcsService) Get(feedID string) (*repository.IcsFeedModel, error) {
	return s.repo.FindFeedByID(feedID)
}

func (s *IcsService) Update(feedID string, input CreateIcsFeedInput) (*repository.IcsFeedModel, error) {
	feed, err := s.repo.FindFeedByID(feedID)
	if err != nil {
		return nil, err
	}

	feed.Name = input.Name
	feed.Description = input.Description
	feed.FilterDays = input.FilterDays
	feed.FilterGroupID = input.FilterGroupID
	feed.FilterType = input.FilterType

	// If switching auth type or updating credentials
	if input.AuthType != feed.AuthType || (input.AuthType == "api_key" && input.ApiKeyID != feed.ApiKeyID) {
		feed.AuthType = input.AuthType
		feed.ApiKeyID = input.ApiKeyID
		if input.AuthType == "basic" {
			// Always derive username from account
			username, _ := s.getAccountUsername(feed.CreatedBy)
			feed.AppUsername = username
			if input.AppPassword != "" {
				hash, _ := bcrypt.GenerateFromPassword([]byte(input.AppPassword), bcrypt.DefaultCost)
				feed.AppPasswordHash = string(hash)
			}
		}
	}

	if err := s.repo.UpdateFeed(feed); err != nil {
		return nil, err
	}
	return feed, nil
}

func (s *IcsService) Delete(feedID string) error {
	return s.repo.DeleteFeed(feedID)
}

// ─── ICS Content Generation ─────────────────────────────────────

const icsProdID = "-//NowAndAgain//Family Calendar//EN"

func (s *IcsService) GenerateICS(token, apiKey, username, password string) (string, error) {
	feed, err := s.repo.FindFeedByToken(token)
	if err != nil {
		return "", fmt.Errorf("feed not found")
	}
	if !feed.Enabled {
		return "", fmt.Errorf("feed disabled")
	}

	// Validate auth
	if !s.validateFeedAccess(feed, apiKey, username, password) {
		return "", fmt.Errorf("authentication required")
	}

	// Get all tasks for this family
	tasks, err := s.taskRepo.ListTasksByFamily(feed.FamilyID)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:" + icsProdID + "\r\n")
	sb.WriteString("X-WR-CALNAME:")
	sb.WriteString(escapeICS(feed.Name))
	sb.WriteString("\r\n")
	sb.WriteString("X-WR-CALDESC:")
	sb.WriteString(escapeICS(feed.Description))
	sb.WriteString("\r\n")
	sb.WriteString("REFRESH-INTERVAL;VALUE=DURATION:PT1H\r\n")

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		// Filter by group
		if feed.FilterGroupID != "" && task.GroupID != "" && task.GroupID != feed.FilterGroupID {
			continue
		}

		// Parse time from schedule data
		schedTime := parseScheduleTime(task.ScheduleType, task.ScheduleData)
		rrule := buildRRule(task.ScheduleType, task.ScheduleData)
		window := scheduleWindow(task.ScheduleType, task.ScheduleData)

		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString("UID:")
		sb.WriteString(task.ID)
		sb.WriteString("\r\n")
		dtStart := schedTime.Format("20060102T150405Z")
		dtEnd := schedTime.Add(window).Format("20060102T150405Z")
		sb.WriteString("DTSTART:")
		sb.WriteString(dtStart)
		sb.WriteString("\r\n")
		sb.WriteString("DTEND:")
		sb.WriteString(dtEnd)
		sb.WriteString("\r\n")
		sb.WriteString("SUMMARY:")
		sb.WriteString(escapeICS(task.Name))
		sb.WriteString("\r\n")
		if rrule != "" {
			sb.WriteString("RRULE:")
			sb.WriteString(rrule)
			sb.WriteString("\r\n")
		}
		sb.WriteString("END:VEVENT\r\n")
	}
	sb.WriteString("END:VCALENDAR\r\n")

	return sb.String(), nil
}

// parseScheduleTime returns the next occurrence time for a task.
func parseScheduleTime(scheduleType, scheduleData string) time.Time {
	var data struct {
		Time string `json:"time"`
		Date string `json:"date"`
	}
	json.Unmarshal([]byte(scheduleData), &data)

	now := timeutil.Now()
	t := data.Time
	if t == "" {
		t = "09:00"
	}

	if scheduleType == "once" && data.Date != "" {
		parsed, err := time.ParseInLocation("2006-01-02 15:04", data.Date+" "+t, time.UTC)
		if err == nil {
			return parsed.UTC()
		}
	}

	h, m := 9, 0
	fmt.Sscanf(t, "%d:%d", &h, &m)
	next := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.UTC)
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// buildRRule returns an iCalendar RRULE string for the given schedule.
func buildRRule(scheduleType, scheduleData string) string {
	switch scheduleType {
	case "daily":
		return "FREQ=DAILY"
	case "weekly":
		var data struct {
			Days []int `json:"days"`
		}
		json.Unmarshal([]byte(scheduleData), &data)
		if len(data.Days) == 0 {
			return "FREQ=WEEKLY"
		}
		// Map 1=Monday...7=Sunday to iCal days MO,TU,WE,TH,FR,SA,SU
		days := []string{"MO", "TU", "WE", "TH", "FR", "SA", "SU"}
		parts := make([]string, 0, len(data.Days))
		for _, d := range data.Days {
			if d >= 1 && d <= 7 {
				parts = append(parts, days[d-1])
			}
		}
		if len(parts) > 0 {
			return "FREQ=WEEKLY;BYDAY=" + strings.Join(parts, ",")
		}
		return "FREQ=WEEKLY"
	case "monthly":
		var data struct {
			Days []int `json:"days"`
		}
		json.Unmarshal([]byte(scheduleData), &data)
		if len(data.Days) == 0 {
			return "FREQ=MONTHLY"
		}
		daysStr := make([]string, len(data.Days))
		for i, d := range data.Days {
			daysStr[i] = fmt.Sprintf("%d", d)
		}
		return "FREQ=MONTHLY;BYMONTHDAY=" + strings.Join(daysStr, ",")
	case "interval":
		var data struct {
			Days int `json:"days"`
		}
		json.Unmarshal([]byte(scheduleData), &data)
		if data.Days <= 0 {
			data.Days = 1
		}
		return fmt.Sprintf("FREQ=DAILY;INTERVAL=%d", data.Days)
	case "once":
		return "" // no recurrence
	}
	return ""
}

// ─── Auth Validation ─────────────────────────────────────────────

func (s *IcsService) validateFeedAccess(feed *repository.IcsFeedModel, apiKey, username, password string) bool {
	switch feed.AuthType {
	case "api_key":
		if apiKey == "" {
			return false
		}
		// Look up and validate the API key
		keys, err := s.apiKeyRepo.ListByUser(feed.CreatedBy)
		if err != nil {
			return false
		}
		for _, k := range keys {
			if k.ID == feed.ApiKeyID && !k.Revoked {
				if hashSHA256(apiKey) == k.KeyHash {
					return true
				}
			}
		}
		return false

	case "basic":
		if username == "" || password == "" {
			return false
		}
		if username != feed.AppUsername {
			return false
		}
		return bcrypt.CompareHashAndPassword([]byte(feed.AppPasswordHash), []byte(password)) == nil
	}
	return false
}

func escapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\r\n", "\\n")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func hashSHA256(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// getAccountUsername returns the login username for the given user ID.
func (s *IcsService) getAccountUsername(userID string) (string, error) {
	acc, err := s.userRepo.FindAccountByUserID(userID)
	if err != nil {
		return "", err
	}
	return acc.Username, nil
}
