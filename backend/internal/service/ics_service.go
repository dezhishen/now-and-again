package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
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

	todos, err := s.taskRepo.ListTodosByFamily(feed.FamilyID, "pending")
	if err != nil {
		return "", err
	}

	now := time.Now()
	cutoff := now.AddDate(0, 0, feed.FilterDays)

	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:" + icsProdID + "\r\n")
	sb.WriteString("X-WR-CALNAME:" + escapeICS(feed.Name) + "\r\n")
	sb.WriteString("X-WR-CALDESC:" + escapeICS(feed.Description) + "\r\n")
	sb.WriteString("REFRESH-INTERVAL;VALUE=DURATION:PT1H\r\n")

	for _, todo := range todos {
		// Filter by date range
		if todo.DueDate.After(cutoff) {
			continue
		}
		// Filter by group
		if feed.FilterGroupID != "" && todo.Task.GroupID != "" && todo.Task.GroupID != feed.FilterGroupID {
			continue
		}
		// Filter by type (personal = only todos with group)
		// For "personal", we'd need the user context. Here "all" means all.
		// Personal filter is handled in the authenticated endpoint variant.

		taskName := todo.Task.Name
		if taskName == "" {
			taskName = "Task"
		}
		summary := taskName
		desc := ""
		if todo.User.DisplayName != "" {
			desc = "Assigned: " + todo.User.DisplayName
		}

		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString("UID:" + todo.ID + "\r\n")
		dtStart := todo.DueStart.Format("20060102T150405Z")
		dtEnd := todo.DueDate.Format("20060102T150405Z")
		sb.WriteString("DTSTART:" + dtStart + "\r\n")
		sb.WriteString("DTEND:" + dtEnd + "\r\n")
		sb.WriteString("SUMMARY:" + escapeICS(summary) + "\r\n")
		if desc != "" {
			sb.WriteString("DESCRIPTION:" + escapeICS(desc) + "\r\n")
		}
		sb.WriteString("STATUS:NEEDS-ACTION\r\n")
		sb.WriteString("END:VEVENT\r\n")
	}
	sb.WriteString("END:VCALENDAR\r\n")

	return sb.String(), nil
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
				// Validate raw key against hash
				if bcrypt.CompareHashAndPassword([]byte(k.KeyHash), []byte(apiKey)) == nil {
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

// getAccountUsername returns the login username for the given user ID.
func (s *IcsService) getAccountUsername(userID string) (string, error) {
	acc, err := s.userRepo.FindAccountByUserID(userID)
	if err != nil {
		return "", err
	}
	return acc.Username, nil
}
