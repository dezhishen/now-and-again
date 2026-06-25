package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ─── chain ───────────────────────────────────────────────────────

var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Manage task chains",
}

var chainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List chains in a family",
	RunE: func(cmd *cobra.Command, args []string) error {
		familyID, _ := cmd.Flags().GetString("family-id")
		fmt.Printf("TODO: list chains for family %s\n", familyID)
		return nil
	},
}

var chainStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a chain (instantiate all steps as tasks)",
	RunE: func(cmd *cobra.Command, args []string) error {
		chainID, _ := cmd.Flags().GetString("chain-id")
		fmt.Printf("TODO: start chain %s\n", chainID)
		return nil
	},
}

func init() {
	chainListCmd.Flags().String("family-id", "", "Family ID (required)")
	chainStartCmd.Flags().String("chain-id", "", "Chain ID (required)")

	chainCmd.AddCommand(chainListCmd)
	chainCmd.AddCommand(chainStartCmd)
}
