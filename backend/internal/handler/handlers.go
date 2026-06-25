package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dezhishen/now-and-again/shared/contracts"
	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// userCtx injects the authenticated user ID into the context.
func userCtx(c *gin.Context) context.Context {
	uid, _ := c.Get("user_id")
	if uid != nil {
		return context.WithValue(c.Request.Context(), "user_id", uid.(string))
	}
	return c.Request.Context()
}

// ─── Response helpers ────────────────────────────────────────────

func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}
func noContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
func badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": msg})
}
func serverError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
}
func paged(c *gin.Context, data interface{}, page, pageSize, total int) {
	totalPages := (total + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"success": true, "data": data,
		"pagination": gin.H{"page": page, "page_size": pageSize, "total": total, "total_pages": totalPages},
	})
}

// ─── Param parsers ───────────────────────────────────────────────

func bindJSON[T any](c *gin.Context) (*T, error) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}
func paramUUID(c *gin.Context, name string) (uuid.UUID, error) {
	return uuid.Parse(c.Param(name))
}
func queryUUID(c *gin.Context, name string) *uuid.UUID {
	s := c.Query(name)
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}
func queryInt(c *gin.Context, name string, defaultVal int) int {
	s := c.Query(name)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
func queryStatus(c *gin.Context, name string) *types.TaskStatus {
	s := c.Query(name)
	if s == "" {
		return nil
	}
	st := types.TaskStatus(s)
	return &st
}
func bodyJSON(c *gin.Context, v interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}

// ─── All handler structs ─────────────────────────────────────────

type AllHandlers struct {
	User         *UserHandlers
	Family       *FamilyHandlers
	SubGroup     *SubGroupHandlers
	Task         *TaskHandlers
	Chain        *ChainHandlers
	Inspection   *InspectionHandlers
	Log          *LogHandlers
	Notification *NotificationHandlers
	ApiKey       *ApiKeyHandlers
}

type UserHandlers struct{ C contracts.UserContract }
type FamilyHandlers struct{ C contracts.FamilyContract }
type SubGroupHandlers struct{ C contracts.SubGroupContract }
type TaskHandlers struct{ C contracts.TaskContract }
type ChainHandlers struct{ C contracts.ChainContract }
type InspectionHandlers struct{ C contracts.InspectionContract }
type LogHandlers struct{ C contracts.LogContract }
type NotificationHandlers struct {
	C contracts.NotificationContract
}
type ApiKeyHandlers struct{ C contracts.ApiKeyContract }

func NewHandlers(all *contracts.AllContracts) *AllHandlers {
	return &AllHandlers{
		User:         &UserHandlers{C: all.User},
		Family:       &FamilyHandlers{C: all.Family},
		SubGroup:     &SubGroupHandlers{C: all.SubGroup},
		Task:         &TaskHandlers{C: all.Task},
		Chain:        &ChainHandlers{C: all.Chain},
		Inspection:   &InspectionHandlers{C: all.Inspection},
		Log:          &LogHandlers{C: all.Log},
		Notification: &NotificationHandlers{C: all.Notification},
		ApiKey:       &ApiKeyHandlers{C: all.ApiKey},
	}
}
