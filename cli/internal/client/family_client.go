package client

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/shared/contracts"
	"github.com/dezhishen/now-and-again/shared/types"
	"github.com/google/uuid"
)

// Compile-time check: FamilyClient must implement FamilyContract.
var _ contracts.FamilyContract = (*FamilyClient)(nil)

// FamilyClient is the HTTP implementation of FamilyContract.
type FamilyClient struct {
	http *HTTPClient
}

func NewFamilyClient(http *HTTPClient) *FamilyClient {
	return &FamilyClient{http: http}
}

// ─── Family CRUD ──────────────────────────────────────────────────

func (c *FamilyClient) Create(ctx context.Context, req *types.CreateFamilyRequest) (*types.Family, error) {
	var f types.Family
	if err := c.http.do("POST", "/api/families", req, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (c *FamilyClient) Join(ctx context.Context, req *types.JoinFamilyRequest) (*types.FamilyMember, error) {
	var m types.FamilyMember
	if err := c.http.do("POST", "/api/families/join", req, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *FamilyClient) Get(ctx context.Context, familyID uuid.UUID) (*types.Family, error) {
	var f types.Family
	if err := c.http.do("GET", fmt.Sprintf("/api/families/%s", familyID), nil, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (c *FamilyClient) ListMyFamilies(ctx context.Context) ([]types.Family, error) {
	var families []types.Family
	if err := c.http.do("GET", "/api/users/me/families", nil, &families); err != nil {
		return nil, err
	}
	return families, nil
}

// ─── Members ──────────────────────────────────────────────────────

func (c *FamilyClient) ListMembers(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error) {
	var members []types.FamilyMember
	if err := c.http.do("GET", fmt.Sprintf("/api/families/%s/members", familyID), nil, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (c *FamilyClient) UpdateMemberRole(ctx context.Context, familyID, userID uuid.UUID, role types.FamilyRole) error {
	req := types.UpdateMemberRoleRequest{Role: role}
	return c.http.do("PUT", fmt.Sprintf("/api/families/%s/members/%s/role", familyID, userID), req, nil)
}

func (c *FamilyClient) RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error {
	return c.http.do("DELETE", fmt.Sprintf("/api/families/%s/members/%s", familyID, userID), nil, nil)
}

func (c *FamilyClient) LeaveFamily(ctx context.Context, familyID uuid.UUID) error {
	return c.http.do("POST", fmt.Sprintf("/api/families/%s/leave", familyID), nil, nil)
}

// ─── Join Requests ────────────────────────────────────────────────

func (c *FamilyClient) ListJoinRequests(ctx context.Context, familyID uuid.UUID) ([]types.FamilyMember, error) {
	var members []types.FamilyMember
	if err := c.http.do("GET", fmt.Sprintf("/api/families/%s/join-requests", familyID), nil, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (c *FamilyClient) ReviewJoinRequest(ctx context.Context, familyID uuid.UUID, req *types.ReviewJoinRequest) error {
	return c.http.do("PUT", fmt.Sprintf("/api/families/%s/join-requests", familyID), req, nil)
}

// ─── Family Group ─────────────────────────────────────────────────

func (c *FamilyClient) CreateGroup(ctx context.Context, familyID uuid.UUID, req *types.CreateFamilyGroupRequest) (*types.FamilyGroup, error) {
	var g types.FamilyGroup
	if err := c.http.do("POST", fmt.Sprintf("/api/families/%s/groups", familyID), req, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func (c *FamilyClient) ListGroups(ctx context.Context, familyID uuid.UUID) ([]types.FamilyGroup, error) {
	var groups []types.FamilyGroup
	if err := c.http.do("GET", fmt.Sprintf("/api/families/%s/groups", familyID), nil, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (c *FamilyClient) JoinGroup(ctx context.Context, groupID uuid.UUID) (*types.FamilyGroupMember, error) {
	var m types.FamilyGroupMember
	if err := c.http.do("POST", fmt.Sprintf("/api/groups/%s/join", groupID), nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *FamilyClient) LeaveGroup(ctx context.Context, groupID uuid.UUID) error {
	return c.http.do("POST", fmt.Sprintf("/api/groups/%s/leave", groupID), nil, nil)
}

func (c *FamilyClient) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]types.FamilyGroupMember, error) {
	var members []types.FamilyGroupMember
	if err := c.http.do("GET", fmt.Sprintf("/api/groups/%s/members", groupID), nil, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (c *FamilyClient) RemoveGroupMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return c.http.do("DELETE", fmt.Sprintf("/api/groups/%s/members/%s", groupID, userID), nil, nil)
}

func (c *FamilyClient) ListGroupJoinRequests(ctx context.Context, groupID uuid.UUID) ([]types.FamilyGroupMember, error) {
	var members []types.FamilyGroupMember
	if err := c.http.do("GET", fmt.Sprintf("/api/groups/%s/join-requests", groupID), nil, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (c *FamilyClient) ReviewGroupJoinRequest(ctx context.Context, groupID uuid.UUID, req *types.ReviewGroupJoinRequest) error {
	return c.http.do("PUT", fmt.Sprintf("/api/groups/%s/join-requests", groupID), req, nil)
}
