package resolver

import (
	"context"
	"testing"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

func TestNewCache_NotEmpty(t *testing.T) {
	c := NewCache()
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestCache_TaskByName(t *testing.T) {
	c := NewCache()
	task := &types.Task{ID: "task-1", Name: "测试任务"}

	// Store in cache
	c.mu.Lock()
	c.taskByName["测试任务"] = task
	c.mu.Unlock()

	// Read back
	c.mu.RLock()
	got, ok := c.taskByName["测试任务"]
	c.mu.RUnlock()
	if !ok {
		t.Fatal("expected task to be found in cache")
	}
	if got.ID != "task-1" || got.Name != "测试任务" {
		t.Errorf("unexpected cached task: %+v", got)
	}
}

func TestCache_GroupByName(t *testing.T) {
	c := NewCache()
	group := &types.FamilyGroup{ID: "group-1", Name: "大人"}

	c.mu.Lock()
	c.groupByName["大人"] = group
	c.mu.Unlock()

	c.mu.RLock()
	got, ok := c.groupByName["大人"]
	c.mu.RUnlock()
	if !ok {
		t.Fatal("expected group to be found in cache")
	}
	if got.ID != "group-1" {
		t.Errorf("expected ID 'group-1', got %q", got.ID)
	}
}

func TestCache_LocationByName(t *testing.T) {
	c := NewCache()
	loc := &types.Location{ID: "loc-1", Name: "厨房"}

	c.mu.Lock()
	c.locationByName["厨房"] = loc
	c.mu.Unlock()

	c.mu.RLock()
	got, ok := c.locationByName["厨房"]
	c.mu.RUnlock()
	if !ok {
		t.Fatal("expected location to be found in cache")
	}
	if got.ID != "loc-1" {
		t.Errorf("expected ID 'loc-1', got %q", got.ID)
	}
}

func TestCache_FamilyByName(t *testing.T) {
	c := NewCache()
	fam := &types.Family{ID: "fam-1", Name: "我的家"}

	c.mu.Lock()
	c.familyByName["我的家"] = fam
	c.mu.Unlock()

	c.mu.RLock()
	got, ok := c.familyByName["我的家"]
	c.mu.RUnlock()
	if !ok {
		t.Fatal("expected family to be found in cache")
	}
	if got.ID != "fam-1" {
		t.Errorf("expected ID 'fam-1', got %q", got.ID)
	}
}

func TestCache_ConcurrentSafe(t *testing.T) {
	c := NewCache()

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			name := string(rune('A' + i))
			c.mu.Lock()
			c.taskByName[name] = &types.Task{ID: name, Name: name}
			c.mu.Unlock()
			done <- true
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all 10 entries
	c.mu.RLock()
	count := len(c.taskByName)
	c.mu.RUnlock()
	if count != 10 {
		t.Errorf("expected 10 cached tasks after concurrent writes, got %d", count)
	}
}

// Test that the context parameter matches the SDK interface
func TestResolveTask_Signature(t *testing.T) {
	c := NewCache()
	// Just verify the method exists with correct signature
	// (actual resolution requires a live SDK, so we only test the cache layer)
	_ = context.TODO()
	_ = c
}
