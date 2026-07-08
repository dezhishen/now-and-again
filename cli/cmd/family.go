package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/spf13/cobra"
)

var familyCmd = &cobra.Command{
	Use:   "family",
	Short: "家庭管理",
	Long:  `创建、加入、查看、切换家庭。`,
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
		action.Printf("Family created: %s (invite code: %s)", f.Name, f.InviteCode)
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
		action.Println("Join request sent")
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

var familySelectCmd = &cobra.Command{
	Use:   "select",
	Short: "切换活跃家庭（支持名称或交互式选择）",
	Long:  `通过 --name 指定名称直接切换，或交互式选择当前活跃的家庭。`,
	Example: `  na family select --name "我的家"
  na family select                # 交互式选择`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		families, err := na.ListMyFamilies(ctx)
		if err != nil {
			return fmt.Errorf("获取家庭列表失败: %w", err)
		}
		if len(families) == 0 {
			return fmt.Errorf("还没有家庭，请先创建: na family create --name \"我的家\"")
		}

		// Try --name flag first
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			for _, fam := range families {
				if fam.Name == name {
					na.SetActiveFamily(fam.ID, fam.Name)
					na.Config().Save()
					action.Printf("✓ 已切换到: %s", fam.Name)
					return nil
				}
			}
			return fmt.Errorf("未找到家庭: %s", name)
		}

		if len(families) == 1 {
			na.SetActiveFamily(families[0].ID, families[0].Name)
			na.Config().Save()
			action.Printf("✓ 已选择唯一家庭: %s", families[0].Name)
			return nil
		}

		fmt.Printf("\n📋 选择活跃家庭:\n\n")
		for i, f := range families {
			mark := " "
			if f.ID == na.ActiveFamilyID() {
				mark = "★"
			}
			fmt.Printf("  %s %d. %s\n", mark, i+1, f.Name)
		}

		fmt.Printf("\n→ 输入编号 (1-%d): ", len(families))
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return nil
		}
		input := strings.TrimSpace(scanner.Text())
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(families) {
			return fmt.Errorf("无效选择: %s", input)
		}

		f := families[n-1]
		na.SetActiveFamily(f.ID, f.Name)
		na.Config().Save()
		action.Printf("✓ 已切换到: %s", f.Name)
		return nil
	},
}

// ─── family leave ─────────────────────────────────────────────────

var familyLeaveCmd = &cobra.Command{
	Use:   "leave",
	Short: "离开当前活跃家庭",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		if err := na.LeaveFamily(context.Background()); err != nil {
			return err
		}
		action.Println("👋 已离开家庭")
		return nil
	},
}

// ─── family members ───────────────────────────────────────────────

var familyMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "列出家庭成员及其角色",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		members, err := na.ListMembers(context.Background())
		if err != nil {
			return err
		}
		if len(members) == 0 {
			action.Println("👤 暂无成员")
			return nil
		}
		fmt.Printf("👤 成员 (%d人):\n\n", len(members))
		for _, m := range members {
			displayName := ""
			if m.User != nil {
				displayName = m.User.DisplayName
			}
			fmt.Printf("  %-12s  %-8s  %s\n", displayName, m.Role, m.UserID[:8])
		}
		fmt.Println()
		return nil
	},
}

// ─── family status ────────────────────────────────────────────────

var familyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前家庭状态（小组、地址、成员）",
	Long: `列出活跃家庭中的所有小组、地址和成员，含 ID 和名称。

JSON 模式返回结构化数据，可供 AI/脚本解析后用于后续 --group-id / --location-id 调用。

示例:
  na family status                  # 表格展示
  na family status -o json          # JSON 输出，含完整 ID`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		ctx := context.Background()

		groups, _ := na.ListGroups(ctx, "")
		locations, _ := na.ListLocations(ctx, "")
		members, _ := na.ListMembers(ctx)

		if action.OutputFormat() != "text" {
			// Structured output: return everything in one envelope
			type statusData struct {
				FamilyName string               `json:"family_name"`
				FamilyID   string               `json:"family_id"`
				Groups     []groupStatusItem    `json:"groups"`
				Locations  []locationStatusItem `json:"locations"`
				Members    []memberStatusItem   `json:"members"`
			}
			data := statusData{
				FamilyName: na.Config().ActiveFamilyName,
				FamilyID:   na.ActiveFamilyID(),
			}
			for _, g := range groups {
				data.Groups = append(data.Groups, groupStatusItem{ID: g.ID, Name: g.Name, Description: g.Description})
			}
			for _, l := range locations {
				data.Locations = append(data.Locations, locationStatusItem{ID: l.ID, Name: l.Name, Kind: l.Kind})
			}
			for _, m := range members {
				displayName := ""
				if m.User != nil {
					displayName = m.User.DisplayName
				}
				data.Members = append(data.Members, memberStatusItem{ID: m.UserID, Name: displayName, Role: string(m.Role)})
			}
			action.PrintSuccess(data, "")
			return nil
		}

		// Text mode: print tables
		fmt.Printf("\n🏠 %s\n", na.Config().ActiveFamilyName)

		if len(groups) > 0 {
			fmt.Printf("\n👥 小组 (%d):\n", len(groups))
			for _, g := range groups {
				fmt.Printf("  %-12s  %s\n", g.Name, g.ID[:8])
			}
		}
		if len(locations) > 0 {
			fmt.Printf("\n📍 地址 (%d):\n", len(locations))
			for _, l := range locations {
				fmt.Printf("  %-12s  %-8s  %s\n", l.Name, l.Kind, l.ID[:8])
			}
		}
		if len(members) > 0 {
			fmt.Printf("\n👤 成员 (%d):\n", len(members))
			for _, m := range members {
				displayName := ""
				if m.User != nil {
					displayName = m.User.DisplayName
				}
				fmt.Printf("  %-12s  %-8s  %s\n", displayName, m.Role, m.UserID[:8])
			}
		}
		fmt.Println()
		return nil
	},
}

// structured output item types
type groupStatusItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
type locationStatusItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind,omitempty"`
}
type memberStatusItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func init() {
	familyCreateCmd.Flags().String("name", "", "家庭名称")
	familyJoinCmd.Flags().String("code", "", "邀请码")
	familySelectCmd.Flags().String("name", "", "家庭名称（直接切换，跳过交互）")

	familyCmd.AddCommand(familyCreateCmd)
	familyCmd.AddCommand(familyJoinCmd)
	familyCmd.AddCommand(familyListCmd)
	familyCmd.AddCommand(familySelectCmd)
	familyCmd.AddCommand(familyStatusCmd)
	familyCmd.AddCommand(familyMembersCmd)
	familyCmd.AddCommand(familyLeaveCmd)
}
