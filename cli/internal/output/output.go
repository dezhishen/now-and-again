// Package output handles formatted CLI output (table, JSON, YAML).
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// Format writes data in the requested format to stdout.
func Format(format string, data interface{}) error {
	switch format {
	case "json":
		return printJSON(data)
	case "table":
		return printTable(data)
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json)", format)
	}
}

func printJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func printTable(data interface{}) error {
	// For now, fall back to pretty-printed JSON with a header.
	// In real implementation, use reflection to build a table.
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STATUS\tID\tTITLE")
	fmt.Fprintln(w, "------\t--\t-----")
	// TODO: iterate over data items
	w.Flush()

	// Fallback
	fmt.Println("(table rendering not yet implemented — showing JSON)")
	return printJSON(data)
}
