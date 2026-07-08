package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/google/uuid"
)

type LocationClient struct {
	http *HTTPClient
}

func NewLocationClient(http *HTTPClient) *LocationClient {
	return &LocationClient{http: http}
}

func (c *LocationClient) Create(ctx context.Context, familyID uuid.UUID, req *types.CreateLocationRequest) (*types.Location, error) {
	var loc types.Location
	if err := c.http.do("POST", "/api/locations", req, &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (c *LocationClient) ListFamilyLocations(ctx context.Context, familyID uuid.UUID, name string) ([]types.Location, error) {
	path := "/api/locations"
	if name != "" {
		path += "?name=" + url.QueryEscape(name)
	}
	var locs []types.Location
	if err := c.http.do("GET", path, nil, &locs); err != nil {
		return nil, err
	}
	return locs, nil
}

func (c *LocationClient) Update(ctx context.Context, locationID uuid.UUID, req *types.UpdateLocationRequest) (*types.Location, error) {
	var loc types.Location
	if err := c.http.do("PUT", fmt.Sprintf("/api/locations/%s", locationID), req, &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (c *LocationClient) Delete(ctx context.Context, locationID uuid.UUID) error {
	return c.http.do("DELETE", fmt.Sprintf("/api/locations/%s", locationID), nil, nil)
}
