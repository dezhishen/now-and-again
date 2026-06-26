package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ─── Context helpers ──────────────────────────────────────────────

func userCtx(c *gin.Context) context.Context {
	uid, _ := c.Get("user_id")
	if uid != nil {
		return context.WithValue(c.Request.Context(), "user_id", uid.(string))
	}
	return c.Request.Context()
}

// ─── Response helpers ─────────────────────────────────────────────

func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}
func badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": msg})
}
func serverError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
}

// ─── Param parsers ────────────────────────────────────────────────

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

// ─── Handler structs ──────────────────────────────────────────────

type AllHandlers struct {
	User      *UserHandlers
	Family    *FamilyHandlers
	ApiKey    *ApiKeyHandlers
	FloorPlan *FloorPlanHandlers
}

type UserHandlers struct{ C contracts.UserContract }
type FamilyHandlers struct{ C contracts.FamilyContract }
type ApiKeyHandlers struct{ C contracts.ApiKeyContract }
type FloorPlanHandlers struct{ C contracts.FloorPlanContract }

func NewHandlers(all *contracts.AllContracts) *AllHandlers {
	return &AllHandlers{
		User:      &UserHandlers{C: all.User},
		Family:    &FamilyHandlers{C: all.Family},
		ApiKey:    &ApiKeyHandlers{C: all.ApiKey},
		FloorPlan: &FloorPlanHandlers{C: all.FloorPlan},
	}
}
