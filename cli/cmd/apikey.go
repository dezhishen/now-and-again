package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var apikeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "API Key 管理",
	Long:  `创建、列出和吊销 API Key。`,
}

var apikeyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新的 API Key",
	Example: `  na apikey create --name "my-tool" --scopes read,write
  na apikey create --name "ci-script" --scopes "task:read,task:write"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name 不能为空")
		}
		scopesStr, _ := cmd.Flags().GetString("scopes")
		scopes := strings.Split(scopesStr, ",")
		for i := range scopes {
			scopes[i] = strings.TrimSpace(scopes[i])
		}

		resp, err := na.ApiKey.Create(context.Background(), &types.CreateApiKeyRequest{
			Name:   name,
			Scopes: scopes,
		})
		if err != nil {
			return err
		}
		if resp.ApiKey != nil && resp.ApiKey.RawKey != "" {
			action.Printf("🔑 API Key 已创建:\n  名称: %s\n  密钥: %s\n  范围: %s",
				resp.ApiKey.Name, resp.ApiKey.RawKey, strings.Join(resp.ApiKey.Scopes, ", "))
			fmt.Println("\n⚠️  请立即保存此密钥，它将不再显示")
		} else {
			action.Println("✅ API Key 已创建")
		}
		return nil
	},
}

var apikeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有 API Key",
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := na.ApiKey.List(context.Background())
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			action.Println("🔑 暂无 API Key")
			return nil
		}
		fmt.Printf("🔑 API Key (%d个):\n\n", len(keys))
		for _, k := range keys {
			exp := "-"
			if k.ExpiresAt != nil && !k.ExpiresAt.IsZero() {
				exp = k.ExpiresAt.Format("2006-01-02")
			}
			s := k.Name
			if s == "" {
				s = k.ID[:8]
			}
			scopes := strings.Join(k.Scopes, ",")
			lastUsed := ""
			if k.LastUsedAt != nil && !k.LastUsedAt.IsZero() {
				lastUsed = fmt.Sprintf("  上次使用: %s", k.LastUsedAt.Format("01-02 15:04"))
			}
			fmt.Printf("  %s  %-16s  %s%s\n", s, scopes, exp, lastUsed)
		}
		fmt.Println()
		return nil
	},
}

var apikeyRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "吊销 API Key",
	Example: "  na apikey revoke --id abc123",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID, _ := cmd.Flags().GetString("id")
		if keyID == "" {
			return fmt.Errorf("--id 不能为空")
		}
		uid, err := uuid.Parse(keyID)
		if err != nil {
			return fmt.Errorf("无效的 ID: %w", err)
		}
		if err := na.ApiKey.Revoke(context.Background(), uid); err != nil {
			return err
		}
		action.Println("🗑️  API Key 已吊销")
		return nil
	},
}

func init() {
	apikeyCreateCmd.Flags().String("name", "", "API Key 名称（必填）")
	apikeyCreateCmd.Flags().String("scopes", "read,write", "权限范围，逗号分隔")

	apikeyRevokeCmd.Flags().String("id", "", "API Key ID")

	apikeyCmd.AddCommand(apikeyCreateCmd)
	apikeyCmd.AddCommand(apikeyListCmd)
	apikeyCmd.AddCommand(apikeyRevokeCmd)
}
