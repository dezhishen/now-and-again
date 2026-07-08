package sdk

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
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

// LeaveFamily leaves the active family.
func (na *NA) LeaveFamily(ctx context.Context) error {
	fid, err := na.requireFamilyID()
	if err != nil {
		return err
	}
	return na.Family.LeaveFamily(ctx, fid)
}

// DeleteFamily archives (soft-deletes) the active family.
func (na *NA) DeleteFamily(ctx context.Context) error {
	fid, err := na.requireFamilyID()
	if err != nil {
		return err
	}
	return na.Family.Delete(ctx, fid)
}

// UpdateFamily updates the active family's properties.
func (na *NA) UpdateFamily(ctx context.Context, req *types.UpdateFamilyRequest) (*types.Family, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Family.Update(ctx, fid, req)
}

// UpdateMemberRole updates a member's role in the active family.
func (na *NA) UpdateMemberRole(ctx context.Context, userID string, role types.FamilyRole) error {
	fid, err := na.requireFamilyID()
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	return na.Family.UpdateMemberRole(ctx, fid, uid, role)
}

// RemoveMember removes a member from the active family.
func (na *NA) RemoveMember(ctx context.Context, userID string) error {
	fid, err := na.requireFamilyID()
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	return na.Family.RemoveMember(ctx, fid, uid)
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

// JoinGroup sends a request to join a group in the active family.
func (na *NA) JoinGroup(ctx context.Context, groupID string) (*types.FamilyGroupMember, error) {
	gid, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group id: %w", err)
	}
	return na.Family.JoinGroup(ctx, gid)
}

// LeaveGroup leaves a group in the active family.
func (na *NA) LeaveGroup(ctx context.Context, groupID string) error {
	gid, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("invalid group id: %w", err)
	}
	return na.Family.LeaveGroup(ctx, gid)
}

// ListGroupMembers returns members of a group.
func (na *NA) ListGroupMembers(ctx context.Context, groupID string) ([]types.FamilyGroupMember, error) {
	gid, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group id: %w", err)
	}
	return na.Family.ListGroupMembers(ctx, gid)
}

// RemoveGroupMember removes a member from a group.
func (na *NA) RemoveGroupMember(ctx context.Context, groupID, userID string) error {
	gid, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("invalid group id: %w", err)
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	return na.Family.RemoveGroupMember(ctx, gid, uid)
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
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Location.ListFamilyLocations(ctx, fid, name)
}

// CreateLocation creates a new location in the active family.
func (na *NA) CreateLocation(ctx context.Context, name, kind, color string, floorPlanID string) (*types.Location, error) {
	fid, err := na.requireFamilyID()
	if err != nil {
		return nil, err
	}
	return na.Location.Create(ctx, fid, &types.CreateLocationRequest{
		Name:        name,
		Kind:        kind,
		Color:       color,
		FloorPlanID: floorPlanID,
	})
}

// UpdateLocation updates a location in the active family.
func (na *NA) UpdateLocation(ctx context.Context, locationID, name, kind, color string) (*types.Location, error) {
	id, err := uuid.Parse(locationID)
	if err != nil {
		return nil, fmt.Errorf("invalid location id: %w", err)
	}
	return na.Location.Update(ctx, id, &types.UpdateLocationRequest{
		Name:  name,
		Kind:  kind,
		Color: color,
	})
}

// DeleteLocation deletes a location from the active family.
func (na *NA) DeleteLocation(ctx context.Context, locationID string) error {
	id, err := uuid.Parse(locationID)
	if err != nil {
		return fmt.Errorf("invalid location id: %w", err)
	}
	return na.Location.Delete(ctx, id)
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
	for _, l := range locations {
		if l.Name == name {
			return &l, nil
		}
	}
	return &locations[0], nil
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
