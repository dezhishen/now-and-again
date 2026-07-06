package sdk

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

// ─── Family operations ────────────────────────────────────────────

// ListMyFamilies returns all families the current user belongs to.
func (na *NA) ListMyFamilies(ctx context.Context) ([]types.Family, error) {
	return na.Family.ListMyFamilies(ctx)
}

// EnsureFamily auto-selects the first available family if none is set.
// Returns the active family ID, or an error if no families exist.
func (na *NA) EnsureFamily(ctx context.Context) (string, error) {
	if na.ActiveFamilyID() != "" {
		return na.ActiveFamilyID(), nil
	}
	families, err := na.Family.ListMyFamilies(ctx)
	if err != nil {
		return "", fmt.Errorf("list families: %w", err)
	}
	if len(families) == 0 {
		return "", fmt.Errorf("no families found — create one with 'na family create --name \"我的家\"'")
	}
	na.SetActiveFamily(families[0].ID, families[0].Name)
	_ = na.Config().Save()
	return families[0].ID, nil
}

// CreateFamily creates a new family and sets it as active.
func (na *NA) CreateFamily(ctx context.Context, name string) (*types.Family, error) {
	f, err := na.Family.Create(ctx, &types.CreateFamilyRequest{Name: name})
	if err != nil {
		return nil, err
	}
	na.SetActiveFamily(f.ID, f.Name)
	return f, nil
}

// JoinFamily joins a family by invite code.
func (na *NA) JoinFamily(ctx context.Context, code string) error {
	_, err := na.Family.Join(ctx, &types.JoinFamilyRequest{InviteCode: code})
	return err
}

// GetFamily returns family details for the active family.
func (na *NA) GetFamily(ctx context.Context) (*types.Family, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Family.Get(ctx, fid)
}

// CreateGroup creates a new group in the active family.
func (na *NA) CreateGroup(ctx context.Context, name, description string) (*types.FamilyGroup, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Family.CreateGroup(ctx, fid, &types.CreateFamilyGroupRequest{
		Name:        name,
		Description: description,
	})
}

// ListGroups returns groups in the active family.
// Set name to a non-empty string to filter by group name (server-side LIKE).
func (na *NA) ListGroups(ctx context.Context, name string) ([]types.FamilyGroup, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Family.ListGroups(ctx, fid, name)
}

// ListMembers returns all members of the active family.
func (na *NA) ListMembers(ctx context.Context) ([]types.FamilyMember, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Family.ListMembers(ctx, fid)
}

// ListLocations returns locations in the active family.
// Set name to a non-empty string to filter by location name (server-side LIKE).
func (na *NA) ListLocations(ctx context.Context, name string) ([]types.Location, error) {
	// Direct HTTP call since we don't have a LocationClient yet.
	var locations []types.Location
	path := "/api/locations"
	if name != "" {
		path += "?name=" + url.QueryEscape(name)
	}
	if err := na.http.Do("GET", path, nil, &locations); err != nil {
		return nil, err
	}
	return locations, nil
}

// CreateLocation creates a new location in the active family.
func (na *NA) CreateLocation(ctx context.Context, name, kind, color string, floorPlanID string) (*types.Location, error) {
	req := &types.CreateLocationRequest{
		Name:  name,
		Kind:  kind,
		Color: color,
	}
	if floorPlanID != "" {
		req.FloorPlanID = floorPlanID
	}
	// Direct HTTP call.
	var loc types.Location
	if err := na.http.Do("POST", "/api/locations", req, &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}

// ─── Name-based queries (read-only) ───────────────────────────────

// FindGroupByName returns the first group in the active family whose name
// matches exactly or contains the given substring. Uses server-side LIKE filtering.
// Returns nil if not found.
func (na *NA) FindGroupByName(ctx context.Context, name string) (*types.FamilyGroup, error) {
	groups, err := na.ListGroups(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}
	// Prefer exact match
	for _, g := range groups {
		if g.Name == name {
			return &g, nil
		}
	}
	// Otherwise return first substring match
	return &groups[0], nil
}

// FindLocationByName returns the first location in the active family whose name
// matches exactly or contains the given substring. Uses server-side LIKE filtering.
// Returns nil if not found.
func (na *NA) FindLocationByName(ctx context.Context, name string) (*types.Location, error) {
	locations, err := na.ListLocations(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(locations) == 0 {
		return nil, nil
	}
	// Prefer exact match
	for _, l := range locations {
		if l.Name == name {
			return &l, nil
		}
	}
	// Otherwise return first substring match
	return &locations[0], nil
}

// FindFamilyByName returns the first family whose name matches exactly
// or contains the given substring. Returns nil if not found.
func (na *NA) FindFamilyByName(ctx context.Context, name string) (*types.Family, error) {
	families, err := na.ListMyFamilies(ctx)
	if err != nil {
		return nil, err
	}
	for _, f := range families {
		if f.Name == name {
			return &f, nil
		}
	}
	for _, f := range families {
		if len(f.Name) >= len(name) && containsSubstringStr(f.Name, name) {
			return &f, nil
		}
	}
	return nil, nil
}

// containsSubstringStr is a string version of containsSubstring.
func containsSubstringStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
