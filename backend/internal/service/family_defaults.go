package service

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/logger"
)

// DefaultFamilyLocations are created when a new family is formed.
// Modify this list to change the defaults for all new families.
var DefaultFamilyLocations = []struct {
	Name  string
	Kind  string
	Color string
}{
	{Name: "厨房", Kind: "indoor", Color: "#f59e0b"},
	{Name: "客厅", Kind: "indoor", Color: "#3b82f6"},
	{Name: "主卧", Kind: "indoor", Color: "#8b5cf6"},
	{Name: "次卧", Kind: "indoor", Color: "#06b6d4"},
	{Name: "卫生间", Kind: "indoor", Color: "#10b981"},
	{Name: "阳台", Kind: "indoor", Color: "#ef4444"},
}

// DefaultFamilyGroups are created when a new family is formed.
var DefaultFamilyGroups = []struct {
	Name        string
	Description string
}{
	{Name: "大人", Description: "家庭成员中的成年人"},
	{Name: "小孩", Description: "家庭成员中的未成年人"},
}

// InitFamilyDefaults creates default locations and groups for a newly created family.
// Errors are logged but not returned — default creation failure should not block family creation.
func InitFamilyDefaults(familyRepo *repository.FamilyRepo, floorPlanRepo *repository.FloorPlanRepo, familyID, userID string) {
	for _, loc := range DefaultFamilyLocations {
		l := &repository.LocationModel{
			FamilyID: familyID,
			Name:     loc.Name,
			Kind:     loc.Kind,
			Color:    loc.Color,
		}
		if err := floorPlanRepo.CreateLocation(l); err != nil {
			logger.Warnf("[family-defaults] create location %s: %v", loc.Name, err)
		}
	}

	for _, grp := range DefaultFamilyGroups {
		g := &repository.FamilyGroupModel{
			FamilyID:    familyID,
			Name:        grp.Name,
			Description: grp.Description,
			CreatedBy:   userID,
		}
		if err := familyRepo.CreateGroup(g); err != nil {
			logger.Warnf("[family-defaults] create group %s: %v", grp.Name, err)
		}
	}

	logger.Infof("[family-defaults] initialized %d locations and %d groups for family %s",
		len(DefaultFamilyLocations), len(DefaultFamilyGroups), familyID)
}
