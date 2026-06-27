package handler

import (
	"context"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/gin-gonic/gin"
)

type rtKey struct{}

func (h *UserHandlers) Register(c *gin.Context) {
	req, err := bindJSON[types.CreateUserRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	user, err := h.C.Register(c.Request.Context(), req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, user)
}

func (h *UserHandlers) Login(c *gin.Context) {
	req, err := bindJSON[types.LoginRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	pair, err := h.C.Login(c.Request.Context(), req)
	if err != nil {
		serverError(c, err)
		return
	}
	setRefreshCookie(c, pair.RefreshToken)
	ok(c, gin.H{
		"access_token": pair.AccessToken,
		"expires_in":   pair.ExpiresIn,
		"user":         pair.User,
	})
}

func (h *UserHandlers) Refresh(c *gin.Context) {
	rt, err := c.Cookie("refresh_token")
	if err != nil || rt == "" {
		badRequest(c, "missing refresh token")
		return
	}
	pair, err := h.C.Refresh(c.Request.Context(), rt)
	if err != nil {
		c.SetCookie("refresh_token", "", -1, "/", "", true, true)
		serverError(c, err)
		return
	}
	setRefreshCookie(c, pair.RefreshToken)
	ok(c, gin.H{
		"access_token": pair.AccessToken,
		"expires_in":   pair.ExpiresIn,
		"user":         pair.User,
	})
}

func (h *UserHandlers) Logout(c *gin.Context) {
	rt, _ := c.Cookie("refresh_token")
	if rt != "" {
		ctx := context.WithValue(c.Request.Context(), rtKey{}, rt)
		c.Request = c.Request.WithContext(ctx)
		_ = h.C.Logout(c.Request.Context())
	}
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	ok(c, gin.H{"message": "logged out"})
}

func (h *UserHandlers) GetMe(c *gin.Context) {
	user, err := h.C.GetMe(userCtx(c))
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, user)
}

func (h *UserHandlers) UpdateMe(c *gin.Context) {
	req, err := bindJSON[types.UpdateUserRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	user, err := h.C.UpdateMe(userCtx(c), req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, user)
}

func (h *UserHandlers) ListUsers(c *gin.Context) {
	users, err := h.C.ListUsers(userCtx(c))
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, users)
}

func setRefreshCookie(c *gin.Context, token string) {
	c.SetCookie("refresh_token", token, int((7 * 24 * time.Hour).Seconds()), "/", "", true, true)
}
