package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ─── inspection ──────────────────────────────────────────────────

var inspectionCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Manage inspections",
}

var inspectionStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new inspection",
	RunE: func(cmd *cobra.Command, args []string) error {
		familyID, _ := cmd.Flags().GetString("family-id")
		title, _ := cmd.Flags().GetString("title")
		fmt.Printf("TODO: start inspection family=%s title=%s\n", familyID, title)
		return nil
	},
}

var inspectionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List inspections in a family",
	RunE: func(cmd *cobra.Command, args []string) error {
		familyID, _ := cmd.Flags().GetString("family-id")
		fmt.Printf("TODO: list inspections for family %s\n", familyID)
		return nil
	},
}

func init() {
	inspectionStartCmd.Flags().String("family-id", "", "Family ID (required)")
	inspectionStartCmd.Flags().String("title", "", "Inspection title (required)")
	inspectionListCmd.Flags().String("family-id", "", "Family ID (required)")

	inspectionCmd.AddCommand(inspectionStartCmd)
	inspectionCmd.AddCommand(inspectionListCmd)
}
