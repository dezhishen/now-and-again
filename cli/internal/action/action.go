// Package action provides actionId generation and structured output formatting
// for CLI commands. Every CLI invocation gets a unique actionId (UUID v4).
//
// Three output modes are supported:
//   - text (default): human-readable with [action:xxx] prefix
//   - json: machine-parseable envelope {"action_id","success","data"}
//   - yaml: same structure as JSON but in YAML
package action

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// ─── Action ID ────────────────────────────────────────────────────

// ActionID is a unique identifier for a single CLI action.
type ActionID string

// NewActionID generates a fresh UUID v4 action ID.
func NewActionID() ActionID {
	return ActionID(uuid.New().String())
}

// Short returns the first 8 characters for display.
func (id ActionID) Short() string {
	s := string(id)
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// ─── Global action state ──────────────────────────────────────────

var (
	currentActionID ActionID
	outputFormat    string // "text", "json", "yaml"
	mu              sync.Mutex
)

// Init sets the current action ID and output format for this CLI invocation.
// If id is empty, a new UUID is generated.
func Init(id, format string) ActionID {
	mu.Lock()
	defer mu.Unlock()
	if id != "" {
		currentActionID = ActionID(id)
	} else {
		currentActionID = NewActionID()
	}
	if format == "" {
		format = "text"
	}
	outputFormat = format
	return currentActionID
}

// CurrentID returns the current action ID.
func CurrentID() ActionID {
	mu.Lock()
	defer mu.Unlock()
	return currentActionID
}

// OutputFormat returns the current output format.
func OutputFormat() string {
	mu.Lock()
	defer mu.Unlock()
	return outputFormat
}

// ─── Output envelope ──────────────────────────────────────────────

// Result is the standard output envelope for all CLI actions.
type Result struct {
	ActionID string `json:"action_id" yaml:"action_id"`
	Success  bool   `json:"success" yaml:"success"`
	Data     any    `json:"data,omitempty" yaml:"data,omitempty"`
	Error    string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ─── Print helpers (text mode) ───────────────────────────────────

// Prefix returns the formatted actionId prefix for text output.
// Example: "[action: a1b2c3d4] "
func Prefix() string {
	return fmt.Sprintf("[action: %s] ", CurrentID().Short())
}

// Printf prints a formatted message with actionId prefix to stdout.
func Printf(format string, args ...interface{}) {
	fmt.Printf("%s%s\n", Prefix(), fmt.Sprintf(format, args...))
}

// Println prints a message with actionId prefix to stdout.
func Println(msg string) {
	fmt.Printf("%s%s\n", Prefix(), msg)
}

// ─── Print helpers (structured mode) ─────────────────────────────

// PrintSuccess outputs a successful result. In text mode, prints the message.
// In JSON/YAML mode, prints the envelope with data.
func PrintSuccess(data any, textMsg string, textArgs ...interface{}) {
	if OutputFormat() == "text" {
		Printf(textMsg, textArgs...)
		return
	}
	writeStructured(Result{
		ActionID: string(CurrentID()),
		Success:  true,
		Data:     data,
	})
}

// PrintError outputs an error result. In text mode, prints the error.
// In JSON/YAML mode, prints the envelope with error.
func PrintError(err error, textPrefix string) {
	if OutputFormat() == "text" {
		fmt.Fprintf(os.Stderr, "%s❌ %s: %v\n", Prefix(), textPrefix, err)
		return
	}
	writeStructured(Result{
		ActionID: string(CurrentID()),
		Success:  false,
		Error:    err.Error(),
	})
}

// PrintTable outputs tabular data. In text mode, prints a formatted table.
// In JSON/YAML mode, prints as structured data.
// headers is a list of column names. rows is a list of string slices.
func PrintTable(headers []string, rows [][]string, emptyMsg string) {
	if OutputFormat() == "text" {
		if len(rows) == 0 && emptyMsg != "" {
			Println(emptyMsg)
			return
		}
		printTextTable(headers, rows)
		return
	}
	// For JSON/YAML, convert to list of maps
	var items []map[string]string
	for _, row := range rows {
		item := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				item[h] = row[i]
			}
		}
		items = append(items, item)
	}
	writeStructured(Result{
		ActionID: string(CurrentID()),
		Success:  true,
		Data:     items,
	})
}

// ─── Structured output writers ───────────────────────────────────

func writeStructured(r Result) {
	switch OutputFormat() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(r)
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		_ = enc.Encode(r)
	}
}

// ─── Text table rendering ────────────────────────────────────────

func printTextTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	// Print header
	printRow(headers, widths)
	// Print separator
	sep := make([]string, len(headers))
	for i, w := range widths {
		sep[i] = repeatStr("─", w)
	}
	printRow(sep, widths)
	// Print rows
	for _, row := range rows {
		printRow(row, widths)
	}
}

func printRow(cells []string, widths []int) {
	for i, cell := range cells {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(padRight(cell, widths[i]))
	}
	fmt.Println()
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + repeatStr(" ", width-len(s))
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
