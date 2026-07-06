package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/dezhishen/now-and-again/backend/pkg/logger"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
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

// ─── Response helpers — all use unified types.APIError ────────────

func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}

func apiError(c *gin.Context, status int, code types.ErrorCode, summary string) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"error":   types.APIError{Code: code, Summary: summary},
	})
}

func apiErrorDetails(c *gin.Context, status int, code types.ErrorCode, summary string, details []types.FieldError) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"error":   types.APIError{Code: code, Summary: summary, Details: details},
	})
}

func badRequest(c *gin.Context, summary string) {
	apiError(c, http.StatusBadRequest, types.ErrBadRequest, summary)
}

func serverError(c *gin.Context, err error) {
	if tryHandleKnownError(c, err) {
		return
	}
	logger.Errorf("handler error: %v", err)
	apiError(c, http.StatusInternalServerError, types.ErrInternal, "服务器内部错误")
}

func tryHandleKnownError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, types.ErrRefreshTokenInvalid) {
		unauthorized(c, "登录状态已过期，请重新登录")
		return true
	}

	if errors.Is(err, types.ErrInvalidInviteCode) {
		badRequest(c, "邀请码无效")
		return true
	}

	msg := strings.ToLower(err.Error())

	if strings.Contains(msg, "not authenticated") || strings.Contains(msg, "authentication required") || strings.Contains(msg, "invalid credentials") {
		unauthorized(c, err.Error())
		return true
	}

	if strings.Contains(msg, "forbidden") || strings.Contains(msg, "permission") || strings.Contains(msg, "only ") || strings.Contains(msg, "not the owner") {
		apiError(c, http.StatusForbidden, types.ErrForbidden, err.Error())
		return true
	}

	if strings.Contains(msg, "not found") {
		notFound(c, err.Error())
		return true
	}

	if strings.Contains(msg, "already") || strings.Contains(msg, "exists") || strings.Contains(msg, "duplicate") || strings.Contains(msg, "conflict") {
		apiError(c, http.StatusConflict, types.ErrConflict, err.Error())
		return true
	}

	if strings.Contains(msg, "invalid") || strings.Contains(msg, "required") || strings.Contains(msg, "must") || strings.Contains(msg, "disabled") || strings.Contains(msg, "参数") || strings.Contains(msg, "无效") || strings.Contains(msg, "无法") {
		badRequest(c, err.Error())
		return true
	}

	return false
}

func unauthorized(c *gin.Context, summary string) {
	apiError(c, http.StatusUnauthorized, types.ErrUnauthorized, summary)
}

func notFound(c *gin.Context, summary string) {
	apiError(c, http.StatusNotFound, types.ErrNotFound, summary)
}

// validationError returns a structured 400 with field-level details.
func validationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make([]types.FieldError, 0, len(ve))
		for _, fe := range ve {
			fields = append(fields, types.FieldError{
				Field:   lowerFirst(fe.Field()),
				Message: validationMsg(fe),
			})
		}
		apiErrorDetails(c, http.StatusBadRequest, types.ErrValidation, "参数校验失败", fields)
		return
	}
	badRequest(c, err.Error())
}

func lowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func validationMsg(fe validator.FieldError) string {
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
