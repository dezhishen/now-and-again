package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/dezhishen/now-and-again/backend/internal/logger"
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

// validationError returns a 400 with field-level validation errors extracted from gin binding.
func validationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make([]gin.H, 0, len(ve))
		for _, fe := range ve {
			fields = append(fields, gin.H{
				"field":   jsonFieldName(fe),
				"message": validationMessage(fe),
				"tag":     fe.Tag(),
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数校验失败", "fields": fields})
		return
	}
	badRequest(c, err.Error())
}

// jsonFieldName extracts the json tag name from a validator.FieldError.
func jsonFieldName(fe validator.FieldError) string {
	// fe.StructField() gives the Go struct field name.
	// fe.StructNamespace() gives the full path like "CreateUserRequest.Username"
	// We just return the last part (field name in snake_case from json tag)
	name := fe.Field()
	return strings.ToLower(name[:1]) + name[1:]
}

// validationMessage returns a human-readable validation error message.
func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "不能为空"
	case "min":
		return "长度不足，最小 " + fe.Param()
	case "max":
		return "超出最大长度限制 " + fe.Param()
	case "email":
		return "邮箱格式不正确"
	default:
		return "校验失败: " + fe.Tag()
	}
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

// bindOrAbort binds JSON and aborts with validation error on failure.
func bindOrAbort[T any](c *gin.Context) (*T, bool) {
	req, err := bindJSON[T](c)
	if err != nil {
		validationError(c, err)
		return nil, false
	}
	return req, true
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
