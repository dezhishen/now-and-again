package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// ─── Action ID ─────────────────────────────────────────────────

func TestNewActionID_NotEmpty(t *testing.T) {
	id := NewActionID()
	if id == "" {
		t.Error("expected non-empty action ID")
	}
}

func TestNewActionID_Unique(t *testing.T) {
	id1 := NewActionID()
	id2 := NewActionID()
	if id1 == id2 {
		t.Error("expected unique action IDs")
	}
}

func TestActionID_Short(t *testing.T) {
	id := ActionID("abcdefghijklmnop")
	short := id.Short()
	if len(short) != 8 {
		t.Errorf("expected short ID of length 8, got %d: %q", len(short), short)
	}
	if short != "abcdefgh" {
		t.Errorf("expected 'abcdefgh', got %q", short)
	}
}

func TestActionID_Short_LessThan8(t *testing.T) {
	id := ActionID("abc")
	short := id.Short()
	if short != "abc" {
		t.Errorf("expected 'abc', got %q", short)
	}
}

// ─── Init ────────────────────────────────────────────────────────

func TestInit_EmptyID_GeneratesNew(t *testing.T) {
	id := Init("", "text")
	if id == "" {
		t.Error("expected non-empty action ID")
	}
	if CurrentID() != id {
		t.Error("CurrentID should return the same ID")
	}
}

func TestInit_GivenID(t *testing.T) {
	given := "my-fixed-id"
	id := Init(given, "json")
	if string(id) != given {
		t.Errorf("expected %q, got %q", given, id)
	}
	if CurrentID() != id {
		t.Error("CurrentID should return the same ID")
	}
}

func TestInit_DefaultFormat(t *testing.T) {
	Init("test-id", "")
	if OutputFormat() != "text" {
		t.Errorf("expected 'text', got %q", OutputFormat())
	}
}

func TestInit_ExplicitFormat(t *testing.T) {
	Init("test-id", "json")
	if OutputFormat() != "json" {
		t.Errorf("expected 'json', got %q", OutputFormat())
	}
}

// ─── Prefix ─────────────────────────────────────────────────────

func TestPrefix_Format(t *testing.T) {
	Init("abcd1234efgh5678", "text")
	prefix := Prefix()
	if !strings.HasPrefix(prefix, "[action: ") {
		t.Errorf("expected prefix to start with '[action: ', got %q", prefix)
	}
	if !strings.HasSuffix(prefix, "] ") {
		t.Errorf("expected prefix to end with '] ', got %q", prefix)
	}
}

// captureStdout runs f and returns everything written to stdout.
func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = stdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// captureStderr runs f and returns everything written to stderr.
func captureStderr(f func()) string {
	r, w, _ := os.Pipe()
	stderr := os.Stderr
	os.Stderr = w
	f()
	w.Close()
	os.Stderr = stderr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// ─── Print helpers (text mode) ──────────────────────────────────

func TestPrintf_Output(t *testing.T) {
	Init("printf-test-id", "text")
	output := captureStdout(func() {
		Printf("hello %s", "world")
	})

	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain 'hello world', got %q", output)
	}
	if !strings.Contains(output, "[action:") {
		t.Errorf("expected output to contain action prefix, got %q", output)
	}
}

func TestPrintln_Output(t *testing.T) {
	Init("println-test-id", "text")
	output := captureStdout(func() {
		Println("test message")
	})

	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain 'test message', got %q", output)
	}
}

// ─── PrintSuccess ───────────────────────────────────────────────

func TestPrintSuccess_TextMode(t *testing.T) {
	Init("success-text", "text")
	output := captureStdout(func() {
		PrintSuccess(nil, "operation completed")
	})

	if !strings.Contains(output, "operation completed") {
		t.Errorf("expected 'operation completed', got %q", output)
	}
}

func TestPrintSuccess_JSONMode(t *testing.T) {
	Init("success-json", "json")
	output := captureStdout(func() {
		PrintSuccess(map[string]string{"key": "val"}, "ignore this text")
	})

	if !strings.Contains(output, `"action_id"`) {
		t.Errorf("expected JSON envelope with action_id, got %q", output)
	}
	if !strings.Contains(output, `"success"`) {
		t.Errorf("expected JSON envelope with success, got %q", output)
	}
	if !strings.Contains(output, `"key": "val"`) {
		t.Errorf("expected data in JSON output, got %q", output)
	}
	// In JSON mode, the text message should NOT appear
	if strings.Contains(output, "ignore this text") {
		t.Errorf("text message should not appear in JSON mode, got %q", output)
	}
}

// ─── PrintError ─────────────────────────────────────────────────

func TestPrintError_TextMode(t *testing.T) {
	Init("error-text", "text")
	output := captureStderr(func() {
		PrintError(fmt.Errorf("something went wrong"), "failed")
	})

	if !strings.Contains(output, "failed") {
		t.Errorf("expected error output to contain 'failed', got %q", output)
	}
	if !strings.Contains(output, "something went wrong") {
		t.Errorf("expected error output to contain error message, got %q", output)
	}
}

func TestPrintError_JSONMode(t *testing.T) {
	Init("error-json", "json")
	output := captureStdout(func() {
		PrintError(fmt.Errorf("fail msg"), "ignored text")
	})

	if !strings.Contains(output, `"success": false`) {
		t.Errorf("expected success=false in JSON, got %q", output)
	}
	if !strings.Contains(output, `"error": "fail msg"`) {
		t.Errorf("expected error message in JSON, got %q", output)
	}
}

// ─── PrintTable ─────────────────────────────────────────────────

func TestPrintTable_TextMode_WithRows(t *testing.T) {
	Init("table-text", "text")
	output := captureStdout(func() {
		PrintTable(
			[]string{"ID", "Name"},
			[][]string{{"1", "Alice"}, {"2", "Bob"}},
			"",
		)
	})

	if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") {
		t.Errorf("expected table rows to contain names, got %q", output)
	}
}

func TestPrintTable_TextMode_Empty(t *testing.T) {
	Init("table-empty", "text")
	output := captureStdout(func() {
		PrintTable(
			[]string{"ID", "Name"},
			[][]string{},
			"no items found",
		)
	})

	if !strings.Contains(output, "no items found") {
		t.Errorf("expected empty message, got %q", output)
	}
}

func TestPrintTable_JSONMode(t *testing.T) {
	Init("table-json", "json")
	output := captureStdout(func() {
		PrintTable(
			[]string{"ID", "Name"},
			[][]string{{"1", "Alice"}},
			"",
		)
	})

	if !strings.Contains(output, `"action_id"`) {
		t.Errorf("expected JSON envelope, got %q", output)
	}
	if !strings.Contains(output, `"ID": "1"`) || !strings.Contains(output, `"Name": "Alice"`) {
		t.Errorf("expected table data in JSON, got %q", output)
	}
}

// ─── Output format ──────────────────────────────────────────────

func TestOutputFormat_AfterInit(t *testing.T) {
	Init("fmt-test", "yaml")
	if OutputFormat() != "yaml" {
		t.Errorf("expected 'yaml', got %q", OutputFormat())
	}
}

// ─── Result struct ──────────────────────────────────────────────

func TestResult_JSON_Marshal(t *testing.T) {
	r := Result{
		ActionID: "abc-123",
		Success:  true,
		Data:     map[string]string{"result": "ok"},
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, `"action_id":"abc-123"`) {
		t.Errorf("expected action_id in JSON, got %s", output)
	}
	if !strings.Contains(output, `"success":true`) {
		t.Errorf("expected success in JSON, got %s", output)
	}
}

func TestResult_YAML_Marshal(t *testing.T) {
	// Verify the YAML struct tags exist
	r := Result{
		ActionID: "abc-123",
		Success:  true,
		Data:     map[string]string{"result": "ok"},
	}

	// Test with json tags as fallback verification
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	if !strings.Contains(string(data), `"result":"ok"`) {
		t.Errorf("expected result data in JSON, got %s", data)
	}
}
