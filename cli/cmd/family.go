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

func init() {
	familyCreateCmd.Flags().String("name", "", "家庭名称")
	familyJoinCmd.Flags().String("code", "", "邀请码")
	familySelectCmd.Flags().String("name", "", "家庭名称（直接切换，跳过交互）")

	familyCmd.AddCommand(familyCreateCmd)
	familyCmd.AddCommand(familyJoinCmd)
	familyCmd.AddCommand(familyListCmd)
	familyCmd.AddCommand(familySelectCmd)
}
