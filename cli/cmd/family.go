package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

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

var familySelectCmd = &cobra.Command{
	Use:   "select",
	Short: "切换活跃家庭（交互式）",
	Long:  `列出所有家庭并让用户选择当前活跃的家庭。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		families, err := na.ListMyFamilies(ctx)
		if err != nil {
			return fmt.Errorf("获取家庭列表失败: %w", err)
		}
		if len(families) == 0 {
			return fmt.Errorf("还没有家庭，请先创建: na family create --name \"我的家\"")
		}
		if len(families) == 1 {
			na.SetActiveFamily(families[0].ID, families[0].Name)
			na.Config().Save()
			fmt.Printf("✓ 已选择唯一家庭: %s\n", families[0].Name)
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
		fmt.Printf("✓ 已切换到: %s\n", f.Name)
		return nil
	},
}

func init() {
	familyCreateCmd.Flags().String("name", "", "家庭名称")
	familyJoinCmd.Flags().String("code", "", "邀请码")

	familyCmd.AddCommand(familyCreateCmd)
	familyCmd.AddCommand(familyJoinCmd)
	familyCmd.AddCommand(familyListCmd)
	familyCmd.AddCommand(familySelectCmd)
}
