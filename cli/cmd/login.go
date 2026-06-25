package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ─── login ───────────────────────────────────────────────────────

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and store a token",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		fmt.Printf("TODO: login as %s\n", username)
		_ = password
		// 1. POST /api/auth/login
		// 2. Store token in ~/.na.yaml
		return nil
	},
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username (required)")
	loginCmd.Flags().StringP("password", "p", "", "Password (required)")
}
