package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
)

type ImageService struct {
	repo         *repository.ImageRepo
	uploadDir    string
	settingsRepo *repository.SettingsRepo
}

func NewImageService(repo *repository.ImageRepo, uploadDir string, settingsRepo *repository.SettingsRepo) *ImageService {
	return &ImageService{repo: repo, uploadDir: uploadDir, settingsRepo: settingsRepo}
}

func (s *ImageService) getStorageType() string {
	setting, err := s.settingsRepo.Get("storage.type")
	if err != nil || setting.Value == "" {
		return "local"
	}
	return setting.Value
}

func (s *ImageService) Save(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*repository.ImageModel, error) {
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext

	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	dst, err := os.Create(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	img := &repository.ImageModel{
		StorageType:  "local",
		FilePath:     filename,
		OriginalName: header.Filename,
		MimeType:     header.Header.Get("Content-Type"),
		Size:         written,
	}
	if err := s.repo.CreateImage(img); err != nil {
		os.Remove(filepath.Join(s.uploadDir, filename))
		return nil, fmt.Errorf("create image record: %w", err)
	}
	return img, nil
}

func (s *ImageService) GetFilePath(ctx context.Context, imageID string) (string, error) {
	img, err := s.repo.FindImageByID(imageID)
	if err != nil {
		return "", fmt.Errorf("image not found")
	}
	return filepath.Join(s.uploadDir, img.FilePath), nil
}
