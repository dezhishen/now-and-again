package service

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/logger"
)

// ── Config (loaded from YAML at startup) ──────────────────────────

type familyDefaultsConfig struct {
	Locations []struct {
		Name  string `yaml:"name"`
		Kind  string `yaml:"kind"`
		Color string `yaml:"color"`
	} `yaml:"locations"`
	Groups []struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	} `yaml:"groups"`
}

var familyDefaults familyDefaultsConfig

// built-in fallback defaults
func builtinFamilyDefaults() familyDefaultsConfig {
	return familyDefaultsConfig{
		Locations: []struct {
			Name  string `yaml:"name"`
			Kind  string `yaml:"kind"`
			Color string `yaml:"color"`
		}{
			{Name: "厨房", Kind: "indoor", Color: "#f59e0b"},
			{Name: "客厅", Kind: "indoor", Color: "#3b82f6"},
			{Name: "主卧", Kind: "indoor", Color: "#8b5cf6"},
			{Name: "次卧", Kind: "indoor", Color: "#06b6d4"},
			{Name: "卫生间", Kind: "indoor", Color: "#10b981"},
			{Name: "阳台", Kind: "indoor", Color: "#ef4444"},
		},
		Groups: []struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		}{
			{Name: "大人", Description: "家庭成员中的成年人"},
			{Name: "小孩", Description: "家庭成员中的未成年人"},
		},
	}
}

// LoadFamilyDefaults reads family default locations/groups from a YAML file.
// Falls back to built-in defaults if the file is missing or invalid.
// Call once at startup.
func LoadFamilyDefaults(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Infof("[family-defaults] no config file at %s, using built-in defaults", path)
		familyDefaults = builtinFamilyDefaults()
		return
	}
	if err := yaml.Unmarshal(data, &familyDefaults); err != nil {
		logger.Warnf("[family-defaults] parse %s: %v, using built-in defaults", path, err)
		familyDefaults = builtinFamilyDefaults()
		return
	}
	logger.Infof("[family-defaults] loaded %d locations, %d groups from %s",
		len(familyDefaults.Locations), len(familyDefaults.Groups), path)
}

// InitFamilyDefaults creates default locations and groups for a newly created family.
// Errors are logged but not returned — default creation failure should not block family creation.
func InitFamilyDefaults(familyRepo *repository.FamilyRepo, floorPlanRepo *repository.FloorPlanRepo, familyID, userID string) {
	for _, loc := range familyDefaults.Locations {
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

	for _, grp := range familyDefaults.Groups {
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
		len(familyDefaults.Locations), len(familyDefaults.Groups), familyID)
}
