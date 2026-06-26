package handler

import (
	"net/http"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

type ImageHandlers struct {
	repo *repository.ImageRepo
}

func NewImageHandlers(repo *repository.ImageRepo) *ImageHandlers {
	return &ImageHandlers{repo: repo}
}

// Serve redirects to the actual image file.
// GET /api/images/:id
func (h *ImageHandlers) Serve(c *gin.Context) {
	imageID := c.Param("id")
	img, err := h.repo.FindImageByID(imageID)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/uploads/"+img.FilePath)
}
