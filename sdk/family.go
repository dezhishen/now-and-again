package sdk

import (
	"context"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

// ─── Family operations ────────────────────────────────────────────

// ListMyFamilies returns all families the current user belongs to.
func (na *NA) ListMyFamilies(ctx context.Context) ([]types.Family, error) {
	return na.Family.ListMyFamilies(ctx)
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

// ListGroups returns all groups in the active family.
func (na *NA) ListGroups(ctx context.Context) ([]types.FamilyGroup, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Family.ListGroups(ctx, fid)
}

// ListMembers returns all members of the active family.
func (na *NA) ListMembers(ctx context.Context) ([]types.FamilyMember, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Family.ListMembers(ctx, fid)
}

// ListLocations returns all locations in the active family.
func (na *NA) ListLocations(ctx context.Context) ([]types.Location, error) {
	// Direct HTTP call since we don't have a LocationClient yet.
	var locations []types.Location
	path := "/api/locations"
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
