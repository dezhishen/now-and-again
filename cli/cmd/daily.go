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

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "查看并处理今天的待办",
	Long: `一键日常：列出待办 → 输入编号 → 输入备注 → 完成。

示例:
  na daily              交互式处理
  na daily --done 3     直接完成第 3 项`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := autoEnsureFamily(); err != nil {
			return err
		}

		todos, err := na.GetPendingTodos(ctx)
		if err != nil {
			return fmt.Errorf("获取待办失败: %w", err)
		}
		if len(todos) == 0 {
			fmt.Println("🎉 当前没有待办事项！")
			return nil
		}

		todoNames := make([]string, len(todos))
		fmt.Printf("\n📋 待办 (%d项):\n\n", len(todos))
		for i, t := range todos {
			name := t.TaskName
			if t.Task != nil {
				name = t.Task.Name
			}
			todoNames[i] = name
			due := ""
			if !t.DueDate.IsZero() {
				due = " ⏰ " + na.FormatTime(t.DueDate, "01-02 15:04")
			}
			fmt.Printf("  %2d. %s%s\n", i+1, name, due)
		}

		directDone, _ := cmd.Flags().GetInt("done")
		if directDone > 0 && directDone <= len(todos) {
			_, err := na.DoneTodo(ctx, todos[directDone-1].ID, "")
			if err != nil {
				return err
			}
			fmt.Printf("✅ 已完成: %s\n", todoNames[directDone-1])
			return nil
		}

		fmt.Printf("\n💡 输入编号完成 (1-%d)，s 跳过当前，q 退出\n", len(todos))
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("\n→ ")
			if !scanner.Scan() {
				break
			}
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}
			switch strings.ToLower(input) {
			case "q", "quit", "exit":
				fmt.Println("👋")
				return nil
			case "s":
				continue
			}
			n, err := strconv.Atoi(input)
			if err != nil || n < 1 || n > len(todos) {
				fmt.Printf("❌ 请输入 1~%d\n", len(todos))
				continue
			}
			remark := ""
			fmt.Print("📝 备注 (回车跳过): ")
			if scanner.Scan() {
				remark = strings.TrimSpace(scanner.Text())
			}
			if _, err := na.DoneTodo(ctx, todos[n-1].ID, remark); err != nil {
				fmt.Println("❌", err)
			} else {
				fmt.Printf("✅ 已完成: %s\n", todoNames[n-1])
			}
		}
		return nil
	},
}

func autoEnsureFamily() error {
	ctx := context.Background()
	if na.ActiveFamilyID() != "" {
		return nil
	}
	families, err := na.ListMyFamilies(ctx)
	if err != nil {
		return fmt.Errorf("获取家庭列表失败: %w", err)
	}
	if len(families) == 0 {
		return fmt.Errorf("还没有家庭，请先创建: na family create --name \"我的家\"")
	}
	na.SetActiveFamily(families[0].ID, families[0].Name)
	na.Config().Save()
	fmt.Printf("🏠 已自动选择家庭: %s\n", families[0].Name)
	return nil
}

func init() {
	dailyCmd.Flags().Int("done", 0, "直接完成第 N 项待办")
}
