package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dezhishen/now-and-again/backend/internal/logger"
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ─── Context helpers ──────────────────────────────────────────────

func userCtx(c *gin.Context) context.Context {
	uid, _ := c.Get("user_id")
	ctx := c.Request.Context()
	if uid != nil {
		ctx = context.WithValue(ctx, "user_id", uid.(string))
	}
	fid, _ := c.Get("family_id")
	if fid != nil {
		ctx = context.WithValue(ctx, "family_id", fid.(string))
	}
	return ctx
}

func familyID(c *gin.Context) string {
	fid, _ := c.Get("family_id")
	if fid != nil {
		return fid.(string)
	}
	return ""
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
	logger.Errorf("handler error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
}
func unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": msg})
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
