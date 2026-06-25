package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ─── family ──────────────────────────────────────────────────────

var familyCmd = &cobra.Command{
	Use:   "family",
	Short: "Manage families",
}

var familyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new family",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		fmt.Printf("TODO: create family name=%s\n", name)
		return nil
	},
}

var familyJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a family by invite code",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetString("code")
		fmt.Printf("TODO: join family with code=%s\n", code)
		return nil
	},
}

func init() {
	familyCreateCmd.Flags().String("name", "", "Family name (required)")
	familyJoinCmd.Flags().String("code", "", "Invite code (required)")

	familyCmd.AddCommand(familyCreateCmd)
	familyCmd.AddCommand(familyJoinCmd)
}
