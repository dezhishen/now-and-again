package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置（交互式或传参）",
	Long: `配置 Now & Again 的连接信息。无需参数时进入交互模式。

交互模式:
  na init                # 交互式问答

传参模式:
  na init -u admin -p 12345678
  na init --key na_key_xxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		// API key mode: no interaction needed.
		if key != "" {
			if err := na.InitWithKey(key); err != nil {
				return fmt.Errorf("初始化失败: %w", err)
			}
			fmt.Println("✓ 已初始化")
			return showConfig()
		}

		// Non-interactive mode: all flags provided.
		if username != "" && password != "" {
			if err := na.Init(username, password); err != nil {
				return fmt.Errorf("初始化失败: %w", err)
			}
			fmt.Println("✓ 初始化成功")
			return showConfig()
		}

		// Interactive mode: no flags → prompt user.
		return interactiveInit()
	},
}

func interactiveInit() error {
	scanner := bufio.NewScanner(os.Stdin)

	// 1. Server URL
	fmt.Print("\n🔗 服务器地址 [http://localhost:8080]: ")
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			na.SetServerURL(input)
		}
	}

	// 2. Username
	fmt.Print("👤 用户名: ")
	if !scanner.Scan() {
		return fmt.Errorf("已取消")
	}
	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	// 3. Password (hidden)
	fmt.Print("🔑 密码: ")
	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("读取密码失败: %w", err)
	}
	fmt.Println()
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}

	// 4. Init
	fmt.Println("\n⏳ 正在登录...")
	if err := na.Init(username, password); err != nil {
		return fmt.Errorf("初始化失败: %w", err)
	}

	// 5. Show active family (auto-selected)
	fmt.Println("\n✓ 初始化成功！")
	showConfig()

	// 6. Offer family selection if multiple families
	families, err := na.ListMyFamilies(context.Background())
	if err == nil && len(families) > 1 {
		fmt.Printf("\n📋 你有 %d 个家庭:\n", len(families))
		for i, f := range families {
			mark := " "
			if f.ID == na.ActiveFamilyID() {
				mark = "★"
			}
			fmt.Printf("  %s %d. %s\n", mark, i+1, f.Name)
		}
		fmt.Print("\n💡 切换家庭: na family select\n")
	}

	return nil
}

func showConfig() error {
	cfg := na.Config()
	fmt.Printf("  Server: %s\n", cfg.ServerURL)
	if cfg.ActiveFamilyName != "" {
		fmt.Printf("  Family: %s\n", cfg.ActiveFamilyName)
	}
	return nil
}

// readPassword reads a password from stdin. Characters are visible for now;
// proper hidden input requires golang.org/x/term (opt-in dependency).
func readPassword() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func init() {
	initCmd.Flags().StringP("username", "u", "", "用户名")
	initCmd.Flags().StringP("password", "p", "", "密码")
	initCmd.Flags().String("key", "", "API Key（跳过登录）")
	initCmd.Flags().Bool("interactive", true, "交互模式")
}
