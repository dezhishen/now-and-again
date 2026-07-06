// Package resolver provides name→ID resolution for CLI commands.
// It uses SDK query methods to find entities by name, with in-memory caching
// to avoid repeated API calls within a single CLI invocation.
//
// Resolution priority: exact name match → substring name match → ID prefix match.
package resolver

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/dezhishen/now-and-again/sdk"
)

// Cache stores resolved entity references in memory for the duration of a CLI run.
type Cache struct {
	mu             sync.RWMutex
	taskByName     map[string]*types.Task
	groupByName    map[string]*types.FamilyGroup
	locationByName map[string]*types.Location
	familyByName   map[string]*types.Family
}

// NewCache creates an empty in-memory cache.
func NewCache() *Cache {
	return &Cache{
		taskByName:     make(map[string]*types.Task),
		groupByName:    make(map[string]*types.FamilyGroup),
		locationByName: make(map[string]*types.Location),
		familyByName:   make(map[string]*types.Family),
	}
}

// ─── Task resolution ──────────────────────────────────────────────

// ResolveTask finds a task by name or ID prefix.
// Priority: exact name → substring name → ID prefix → full UUID.
func (c *Cache) ResolveTask(ctx context.Context, na *sdk.NA, input string) (*types.Task, error) {
	// Check cache
	c.mu.RLock()
	if t, ok := c.taskByName[input]; ok {
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	// 1. Try by name (exact then substring)
	t, err := na.FindTaskByName(ctx, input)
	if err != nil {
		return nil, err
	}
	if t != nil {
		c.mu.Lock()
		c.taskByName[input] = t
		c.mu.Unlock()
		return t, nil
	}

	// 2. Try by ID prefix
	t, err = na.ResolveTaskID(ctx, input)
	if err != nil {
		// Wrap as a more descriptive error
		return nil, fmt.Errorf("未找到匹配的任务: %s", input)
	}
	c.mu.Lock()
	c.taskByName[input] = t
	c.mu.Unlock()
	return t, nil
}

// ResolveTaskID resolves to a plain UUID string.
func (c *Cache) ResolveTaskID(ctx context.Context, na *sdk.NA, input string) (string, error) {
	t, err := c.ResolveTask(ctx, na, input)
	if err != nil {
		return "", err
	}
	return t.ID, nil
}

// ─── Group resolution ─────────────────────────────────────────────

// ResolveGroup finds a group by name.
// Priority: exact name → substring name.
func (c *Cache) ResolveGroup(ctx context.Context, na *sdk.NA, input string) (*types.FamilyGroup, error) {
	c.mu.RLock()
	if g, ok := c.groupByName[input]; ok {
		c.mu.RUnlock()
		return g, nil
	}
	c.mu.RUnlock()

	// Use server-side name filtering first
	groups, err := na.ListGroups(ctx, input)
	if err != nil {
		return nil, err
	}
	// Exact name match
	for _, g := range groups {
		if g.Name == input {
			c.mu.Lock()
			c.groupByName[input] = &g
			c.mu.Unlock()
			return &g, nil
		}
	}
	// Substring name match
	var substringMatches []types.FamilyGroup
	inputLower := strings.ToLower(input)
	for _, g := range groups {
		if strings.Contains(strings.ToLower(g.Name), inputLower) {
			substringMatches = append(substringMatches, g)
		}
	}
	if len(substringMatches) == 1 {
		c.mu.Lock()
		c.groupByName[input] = &substringMatches[0]
		c.mu.Unlock()
		return &substringMatches[0], nil
	}
	if len(substringMatches) > 1 {
		return nil, fmt.Errorf("找到多个匹配的小组 %q: %s", input, joinGroupNames(substringMatches))
	}

	// Try by ID prefix
	inputLower = strings.ToLower(input)
	for _, g := range groups {
		if strings.HasPrefix(strings.ToLower(g.ID), inputLower) {
			c.mu.Lock()
			c.groupByName[input] = &g
			c.mu.Unlock()
			return &g, nil
		}
	}

	return nil, fmt.Errorf("未找到匹配的小组: %s", input)
}

// ResolveGroupID resolves to a plain UUID string.
func (c *Cache) ResolveGroupID(ctx context.Context, na *sdk.NA, input string) (string, error) {
	g, err := c.ResolveGroup(ctx, na, input)
	if err != nil {
		return "", err
	}
	return g.ID, nil
}

// ─── Location resolution ──────────────────────────────────────────

// ResolveLocation finds a location by name.
// Priority: exact name → substring name → ID prefix.
func (c *Cache) ResolveLocation(ctx context.Context, na *sdk.NA, input string) (*types.Location, error) {
	c.mu.RLock()
	if l, ok := c.locationByName[input]; ok {
		c.mu.RUnlock()
		return l, nil
	}
	c.mu.RUnlock()

	locations, err := na.ListLocations(ctx, input)
	if err != nil {
		return nil, err
	}
	// Exact name match
	for _, l := range locations {
		if l.Name == input {
			c.mu.Lock()
			c.locationByName[input] = &l
			c.mu.Unlock()
			return &l, nil
		}
	}
	// Substring name match
	var substringMatches []types.Location
	inputLower := strings.ToLower(input)
	for _, l := range locations {
		if strings.Contains(strings.ToLower(l.Name), inputLower) {
			substringMatches = append(substringMatches, l)
		}
	}
	if len(substringMatches) == 1 {
		c.mu.Lock()
		c.locationByName[input] = &substringMatches[0]
		c.mu.Unlock()
		return &substringMatches[0], nil
	}
	if len(substringMatches) > 1 {
		return nil, fmt.Errorf("找到多个匹配的地址 %q: %s", input, joinLocationNames(substringMatches))
	}

	// Try by ID prefix
	for _, l := range locations {
		if strings.HasPrefix(strings.ToLower(l.ID), inputLower) {
			c.mu.Lock()
			c.locationByName[input] = &l
			c.mu.Unlock()
			return &l, nil
		}
	}

	return nil, fmt.Errorf("未找到匹配的地址: %s", input)
}

// ResolveLocationID resolves to a plain UUID string.
func (c *Cache) ResolveLocationID(ctx context.Context, na *sdk.NA, input string) (string, error) {
	l, err := c.ResolveLocation(ctx, na, input)
	if err != nil {
		return "", err
	}
	return l.ID, nil
}

// ─── Family resolution ────────────────────────────────────────────

// ResolveFamily finds a family by name.
func (c *Cache) ResolveFamily(ctx context.Context, na *sdk.NA, input string) (*types.Family, error) {
	c.mu.RLock()
	if f, ok := c.familyByName[input]; ok {
		c.mu.RUnlock()
		return f, nil
	}
	c.mu.RUnlock()

	f, err := na.FindFamilyByName(ctx, input)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, fmt.Errorf("未找到匹配的家庭: %s", input)
	}
	c.mu.Lock()
	c.familyByName[input] = f
	c.mu.Unlock()
	return f, nil
}

// ─── Helpers ──────────────────────────────────────────────────────

func joinGroupNames(groups []types.FamilyGroup) string {
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Name
	}
	return strings.Join(names, ", ")
}

func joinLocationNames(locations []types.Location) string {
	names := make([]string, len(locations))
	for i, l := range locations {
		names[i] = l.Name
	}
	return strings.Join(names, ", ")
}
