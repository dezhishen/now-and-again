package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"

	"github.com/google/uuid"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

// ─── Floor Plan ──────────────────────────────────────────────────

func (s *FloorPlanService) Upload(ctx context.Context, familyID uuid.UUID, label string, isCover bool, file multipart.File, header *multipart.FileHeader) (*types.FloorPlan, error) {
	if label == "" {
		label = "1F"
	}

	// Save image via image service
	img, err := s.imageSvc.Save(ctx, file, header)
	if err != nil {
		return nil, fmt.Errorf("save image: %w", err)
	}

	// If this is the cover, clear existing covers
	if isCover {
		s.repo.ClearCoverForFamily(familyID.String())
	}
	// If it's the first floor plan, auto-set as cover
	existing, _ := s.repo.ListFloorPlansByFamilyID(familyID.String())
	if len(existing) == 0 {
		isCover = true
	}

	fp := &repository.FloorPlanModel{
		FamilyID: familyID.String(),
		Label:    label,
		ImageID:  img.ID,
		IsCover:  isCover,
	}
	if err := s.repo.CreateFloorPlan(fp); err != nil {
		return nil, fmt.Errorf("create floor plan: %w", err)
	}

	result := floorPlanModelToType(fp)
	result.ImageURL = "/api/images/" + img.ID
	return &result, nil
}

func (s *FloorPlanService) ListByFamily(ctx context.Context, familyID uuid.UUID) ([]types.FloorPlan, error) {
	plans, err := s.repo.ListFloorPlansByFamilyID(familyID.String())
	if err != nil {
		return nil, err
	}
	result := make([]types.FloorPlan, len(plans))
	for i, fp := range plans {
		result[i] = floorPlanModelToType(&fp)
	}
	return result, nil
}

func (s *FloorPlanService) GetByID(ctx context.Context, planID uuid.UUID) (*types.FloorPlan, error) {
	fp, err := s.repo.FindFloorPlanByID(planID.String())
	if err != nil {
		return nil, fmt.Errorf("floor plan not found")
	}

	locs, _ := s.repo.ListLocations(fp.ID)
	locList := make([]types.Location, len(locs))
	for i, l := range locs {
		locList[i] = locationModelToType(&l)
	}

	result := floorPlanModelToType(fp)
	result.Locations = locList
	return &result, nil
}

func (s *FloorPlanService) Delete(ctx context.Context, planID uuid.UUID) error {
	fp, err := s.repo.FindFloorPlanByID(planID.String())
	if err != nil {
		return fmt.Errorf("floor plan not found")
	}
	// Delete image file from disk
	if fp.Image.ID != "" {
		imgPath, _ := s.imageSvc.GetFilePath(ctx, fp.ImageID)
		os.Remove(imgPath)
		s.imageRepo.DeleteImage(fp.ImageID)
	}
	return s.repo.DeleteFloorPlanByID(planID.String())
}

func (s *FloorPlanService) SetCover(ctx context.Context, planID uuid.UUID) error {
	fp, err := s.repo.FindFloorPlanByID(planID.String())
	if err != nil {
		return fmt.Errorf("floor plan not found")
	}
	s.repo.ClearCoverForFamily(fp.FamilyID)
	return s.repo.SetCover(planID.String())
}

// ─── Rooms ───────────────────────────────────────────────────────

func (s *FloorPlanService) CreateLocation(ctx context.Context, floorPlanID uuid.UUID, req *types.CreateLocationRequest) (*types.Location, error) {
	color := req.Color
	if color == "" {
		color = "#3b82f6"
	}
	loc := &repository.LocationModel{
		FloorPlanID: floorPlanID.String(),
		Name:        req.Name,
		PointX:      req.Point.X,
		PointY:      req.Point.Y,
		Color:       color,
	}
	if err := s.repo.CreateLocation(loc); err != nil {
		return nil, fmt.Errorf("create location: %w", err)
	}
	result := locationModelToType(loc)
	return &result, nil
}

func (s *FloorPlanService) ListLocations(ctx context.Context, floorPlanID uuid.UUID) ([]types.Location, error) {
	locs, err := s.repo.ListLocations(floorPlanID.String())
	if err != nil {
		return nil, err
	}
	result := make([]types.Location, len(locs))
	for i, l := range locs {
		result[i] = locationModelToType(&l)
	}
	return result, nil
}

func (s *FloorPlanService) UpdateLocation(ctx context.Context, locationID uuid.UUID, req *types.UpdateLocationRequest) (*types.Location, error) {
	l, err := s.repo.FindLocationByID(locationID.String())
	if err != nil {
		return nil, fmt.Errorf("location not found")
	}
	if req.Name != "" {
		l.Name = req.Name
	}
	if req.Point != nil {
		l.PointX = req.Point.X
		l.PointY = req.Point.Y
	}
	if req.Color != "" {
		l.Color = req.Color
	}
	if err := s.repo.UpdateLocation(l); err != nil {
		return nil, fmt.Errorf("update location: %w", err)
	}
	result := locationModelToType(l)
	return &result, nil
}

func (s *FloorPlanService) DeleteLocation(ctx context.Context, locationID uuid.UUID) error {
	return s.repo.DeleteLocation(locationID.String())
}

// ─── Helpers ─────────────────────────────────────────────────────

func locationModelToType(l *repository.LocationModel) types.Location {
	return types.Location{
		ID:          l.ID,
		FloorPlanID: l.FloorPlanID,
		Name:        l.Name,
		Point:       types.Point{X: l.PointX, Y: l.PointY},
		Color:       l.Color,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

func floorPlanModelToType(fp *repository.FloorPlanModel) types.FloorPlan {
	imageURL := ""
	if fp.Image.ID != "" {
		imageURL = "/api/images/" + fp.Image.ID
	}
	return types.FloorPlan{
		ID: fp.ID, FamilyID: fp.FamilyID, Label: fp.Label,
		ImageURL: imageURL, IsCover: fp.IsCover,
		Width: fp.Width, Height: fp.Height,
		CreatedAt: fp.CreatedAt, UpdatedAt: fp.UpdatedAt,
	}
}
