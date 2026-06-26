package handler

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

type SettingsHandlers struct {
	repo *repository.SettingsRepo
}

func NewSettingsHandlers(repo *repository.SettingsRepo) *SettingsHandlers {
	return &SettingsHandlers{repo: repo}
}

func (h *SettingsHandlers) GetAll(c *gin.Context) {
	settings, err := h.repo.GetAll()
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, settings)
}

func (h *SettingsHandlers) Update(c *gin.Context) {
	var updates map[string]string
	if err := c.ShouldBindJSON(&updates); err != nil {
		badRequest(c, "invalid body")
		return
	}
	for k, v := range updates {
		if err := h.repo.Set(k, v); err != nil {
			serverError(c, err)
			return
		}
	}
	ok(c, gin.H{"message": "settings updated"})
}
