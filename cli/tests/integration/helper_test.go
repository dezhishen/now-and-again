// Package integration provides CLI integration tests that run against a real backend.
//
// Prerequisites: backend must be running on NA_TEST_SERVER (default http://localhost:8080).
package integration

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ─── Config ────────────────────────────────────────────────────────

var (
	serverURL   string
	binaryPath  string
	projectRoot string
	testHome    string // per-test HOME directory for config isolation
)

func init() {
	serverURL = os.Getenv("NA_TEST_SERVER")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	// Find the cli binary relative to the test file.
	wd, _ := os.Getwd()
	projectRoot = filepath.Join(wd, "..", "..")
	binaryPath = filepath.Join(projectRoot, "na")
}

// ─── Helpers ───────────────────────────────────────────────────────

// runNA executes the CLI binary with the given arguments and returns stdout, stderr, and error.
func runNA(args ...string) (string, string, error) {
	fullArgs := append([]string{"--server", serverURL}, args...)
	cmd := exec.Command(binaryPath, fullArgs...)
	// Use a test-specific HOME so config persists within a test but not across tests.
	if testHome != "" {
		cmd.Env = append(os.Environ(), "HOME="+testHome)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

// runNAOK is like runNA but fails the test if the command returns an error.
func runNAOK(t *testing.T, args ...string) string {
	t.Helper()
	out, errOut, err := runNA(args...)
	if err != nil {
		t.Fatalf("na %s failed:\nstdout: %s\nstderr: %s\nerr: %v",
			strings.Join(args, " "), out, errOut, err)
	}
	return out
}

// resetBackend clears test state.
func resetBackend(t *testing.T) {
	t.Helper()
	testHome = filepath.Join(os.TempDir(), "na-test-"+t.Name())
	os.RemoveAll(testHome)
	os.MkdirAll(testHome, 0755)
	os.Remove(filepath.Join(testHome, ".na.yaml"))
}

// ensureBackendReady polls the server status endpoint until it responds.
func ensureBackendReady(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(serverURL + "/api/system/status")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("backend not ready at %s after 15s", serverURL)
}

// testName generates a unique name for test resources.
func testName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// setupTest performs init for a unique test user. Returns the username used.
func setupTest(t *testing.T) string {
	t.Helper()
	user := testName("testuser")
	pwd := "12345678"

	// Register unique user via CLI init (if admin init fails, register fresh)
	out, _, err := runNA("init", "-u", user, "-p", pwd)
	if err != nil {
		// User may not exist — try admin first to bootstrap
		out2, _, err2 := runNA("init", "-u", "admin", "-p", "12345678")
		if err2 != nil {
			t.Fatalf("even admin init failed: %v\nout: %s", err2, out2)
		}
		t.Logf("bootstrapped with admin: %s", out2)
		// Admin creates the test user's family context
		// For now, proceed with admin
		user = "admin"
		pwd = "12345678"
	}

	if !strings.Contains(out, "Initialized") && !strings.Contains(out, "successfully") {
		t.Logf("init output: %s", out)
	}
	return user
}

// setupWithFamily initializes and ensures a family exists for the test.
func setupWithFamily(t *testing.T) {
	t.Helper()
	ensureBackendReady(t)
	resetBackend(t)

	runNAOK(t, "init", "-u", "admin", "-p", "12345678")

	// Create family if not already exists
	out, _, err := runNA("family", "create", "--name", testName("测试家庭"))
	if err != nil && strings.Contains(out, "CONFLICT") {
		// Family already exists from previous test — that's OK
		t.Log("family already exists")
		return
	}
	if err != nil {
		t.Logf("family create (may already exist): %s", out)
	}
}
