package handler

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type IcsHandlers struct {
	Svc *service.IcsService
}

// ─── Authenticated endpoints ─────────────────────────────────────

func (h *IcsHandlers) Create(c *gin.Context) {
	req, err := bindJSON[createIcsReq](c)
	if err != nil {
		badRequest(c, "invalid request")
		return
	}
	familyID := familyID(c)
	userID := c.GetString("user_id")

	input := service.CreateIcsFeedInput{
		Name:          req.Name,
		Description:   req.Description,
		FilterDays:    req.FilterDays,
		FilterGroupID: req.FilterGroupID,
		FilterType:    req.FilterType,
		AuthType:      req.AuthType,
		ApiKeyID:      req.ApiKeyID,
		AppPassword:   req.AppPassword,
		FamilyID:      familyID,
		CreatedBy:     userID,
	}
	feed, err := h.Svc.Create(input)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, feedToResponse(feed))
}

func (h *IcsHandlers) List(c *gin.Context) {
	familyID := familyID(c)
	feeds, err := h.Svc.List(familyID)
	if err != nil {
		serverError(c, err)
		return
	}
	result := make([]icsFeedResponse, len(feeds))
	for i, f := range feeds {
		result[i] = feedToResponse(&f)
	}
	ok(c, result)
}

func (h *IcsHandlers) Get(c *gin.Context) {
	feedID := c.Param("feed_id")
	feed, err := h.Svc.Get(feedID)
	if err != nil {
		badRequest(c, "feed not found")
		return
	}
	ok(c, feedToResponse(feed))
}

func (h *IcsHandlers) Update(c *gin.Context) {
	req, err := bindJSON[createIcsReq](c)
	if err != nil {
		badRequest(c, "invalid request")
		return
	}
	feedID := c.Param("feed_id")
	input := service.CreateIcsFeedInput{
		Name:          req.Name,
		Description:   req.Description,
		FilterDays:    req.FilterDays,
		FilterGroupID: req.FilterGroupID,
		FilterType:    req.FilterType,
		AuthType:      req.AuthType,
		ApiKeyID:      req.ApiKeyID,
		AppPassword:   req.AppPassword,
	}
	feed, err := h.Svc.Update(feedID, input)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, feedToResponse(feed))
}

func (h *IcsHandlers) Delete(c *gin.Context) {
	feedID := c.Param("feed_id")
	if err := h.Svc.Delete(feedID); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"deleted": true})
}

// ─── Public ICS endpoint (bypasses JWT, uses token + optional extra auth) ───

func (h *IcsHandlers) ServeICS(c *gin.Context) {
	token := c.Param("token")
	if len(token) > 4 && token[len(token)-4:] == ".ics" {
		token = token[:len(token)-4]
	}

	apiKey := c.Query("key")
	username, password, _ := c.Request.BasicAuth()

	// Look up the feed to determine its auth type before validating.
	// This is needed so we can send a proper WWW-Authenticate challenge
	// for Basic Auth feeds — required by Thunderbird and other calendar clients.
	feed, feedErr := h.Svc.FindFeedByToken(token)
	if feedErr == nil && feed.AuthType == "basic" && username == "" {
		c.Header("WWW-Authenticate", `Basic realm="Now And Again Calendar", charset="UTF-8"`)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ics, err := h.Svc.GenerateICS(token, apiKey, username, password)
	if err != nil {
		// For basic auth feeds, include WWW-Authenticate so clients can re-prompt
		// for credentials (e.g. when the password was wrong).
		if feedErr == nil && feed.AuthType == "basic" {
			c.Header("WWW-Authenticate", `Basic realm="Now And Again Calendar", charset="UTF-8"`)
		}
		unauthorized(c, err.Error())
		return
	}

	c.Header("Content-Type", "text/calendar; charset=utf-8")
	// Use feed name as filename so calendar clients show a friendly name
	// instead of the UUID token.
	filename := "calendar.ics"
	if feedErr == nil && feed.Name != "" {
		filename = sanitizeICSFilename(feed.Name)
	}
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.String(200, ics)
}

// ─── Types ───────────────────────────────────────────────────────

type createIcsReq struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description,omitempty"`
	FilterDays    int    `json:"filter_days,omitempty"`
	FilterGroupID string `json:"filter_group_id,omitempty"`
	FilterType    string `json:"filter_type,omitempty"`
	AuthType      string `json:"auth_type,omitempty"`
	ApiKeyID      string `json:"api_key_id,omitempty"`
	AppPassword   string `json:"app_password,omitempty"`
}

type icsFeedResponse struct {
	ID            string `json:"id"`
	FamilyID      string `json:"family_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	FilterDays    int    `json:"filter_days"`
	FilterGroupID string `json:"filter_group_id,omitempty"`
	FilterType    string `json:"filter_type"`
	AuthType      string `json:"auth_type"`
	AppUsername   string `json:"app_username,omitempty"`
	ApiKeyPrefix  string `json:"api_key_prefix,omitempty"`
	IcsURL        string `json:"ics_url"`
	Enabled       bool   `json:"enabled"`
	CreatedAt     string `json:"created_at"`
}

func feedToResponse(f *repository.IcsFeedModel) icsFeedResponse {
	slug := slugify(f.Name)
	if slug == "" {
		slug = "calendar"
	}
	r := icsFeedResponse{
		ID:            f.ID,
		FamilyID:      f.FamilyID,
		Name:          f.Name,
		Description:   f.Description,
		FilterDays:    f.FilterDays,
		FilterGroupID: f.FilterGroupID,
		FilterType:    f.FilterType,
		AuthType:      f.AuthType,
		AppUsername:   f.AppUsername,
		Enabled:       f.Enabled,
		IcsURL:        "/api/ics/" + f.AccessToken + "/" + slug + ".ics",
		CreatedAt:     f.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if f.ApiKey != nil {
		r.ApiKeyPrefix = f.ApiKey.KeyPrefix
	}
	return r
}

// sanitizeICSFilename returns a safe filename for Content-Disposition header.
func sanitizeICSFilename(name string) string {
	// Remove characters unsafe for filenames
	safe := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return -1
		}
		return r
	}, name)
	if safe == "" {
		return "calendar.ics"
	}
	return safe + ".ics"
}

// slugify converts a feed name to a URL-safe slug for the ICS URL path.
// e.g. "My Family Calendar" → "My-Family-Calendar"
func slugify(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' || r == '.' {
			b.WriteRune('-')
		}
		// CJK and other special chars are dropped for URL safety
	}
	result := strings.Trim(b.String(), "-")
	return result
}
