package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// ─── family ──────────────────────────────────────────────────────

var familyCmd = &cobra.Command{
	Use:   "family",
	Short: "Manage families",
	Long:  `Create, join, and list your families.`,
}

var familyCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new family",
	Example: "  na family create --name \"我的家\"",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		f, err := na.CreateFamily(context.Background(), name)
		if err != nil {
			return err
		}
		fmt.Printf("Family created: %s (invite code: %s)\n", f.Name, f.InviteCode)
		return nil
	},
}

var familyJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a family by invite code",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetString("code")
		if code == "" {
			return fmt.Errorf("--code is required")
		}
		if err := na.JoinFamily(context.Background(), code); err != nil {
			return err
		}
		fmt.Println("Join request sent")
		return nil
	},
}

var familyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List my families",
	RunE: func(cmd *cobra.Command, args []string) error {
		families, err := na.ListMyFamilies(context.Background())
		if err != nil {
			return err
		}
		if len(families) == 0 {
			fmt.Println("No families")
			return nil
		}
		for _, f := range families {
			fmt.Printf("  %s  %s  (code: %s)\n", f.ID[:8], f.Name, f.InviteCode)
		}
		return nil
	},
}

func init() {
	familyCreateCmd.Flags().String("name", "", "Family name (required)")
	familyJoinCmd.Flags().String("code", "", "Invite code (required)")

	familyCmd.AddCommand(familyCreateCmd)
	familyCmd.AddCommand(familyJoinCmd)
	familyCmd.AddCommand(familyListCmd)
}
