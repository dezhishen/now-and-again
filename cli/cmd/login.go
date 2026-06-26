package cmd

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/spf13/cobra"
)

// ─── login ───────────────────────────────────────────────────────

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and get a token",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		if username == "" || password == "" {
			return fmt.Errorf("--username and --password are required")
		}

		pair, err := allClients.User.Login(context.Background(), &types.LoginRequest{
			Username: username,
			Password: password,
		})
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		fmt.Printf("Login successful!\n")
		fmt.Printf("Access Token: %s\n", pair.AccessToken)
		fmt.Printf("Expires in: %d seconds\n", pair.ExpiresIn)
		if pair.User != nil {
			fmt.Printf("User: %s (%s)\n", pair.User.DisplayName, pair.User.Email)
		}

		fmt.Println("\nTip: export NA_TOKEN=<access_token> to use it with other commands")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username (required)")
	loginCmd.Flags().StringP("password", "p", "", "Password (required)")
}
