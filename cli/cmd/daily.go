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

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "查看并处理今天的待办",
	Long: `查看今天的待办事项。

交互模式 (TTY):
  na daily              交互式处理（输入编号 → 备注 → 完成）

非交互模式 (AI/脚本):
  na daily              仅列出待办（自动检测非TTY）
  na daily --done 3     直接完成第 3 项
  na daily --output json JSON格式输出`,
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
			action.Println("🎉 当前没有待办事项！")
			return nil
		}

		todoNames := make([]string, len(todos))
		// Pre-load groups and locations for display
		groups, _ := na.ListGroups(ctx, "")
		locations, _ := na.ListLocations(ctx, "")
		groupMap := make(map[string]string)
		for _, g := range groups {
			groupMap[g.ID] = g.Name
		}
		locMap := make(map[string]string)
		for _, l := range locations {
			locMap[l.ID] = l.Name
		}

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
			extra := ""
			if t.Task != nil {
				if gn, ok := groupMap[t.Task.GroupID]; ok {
					extra += fmt.Sprintf(" [%s]", gn)
				}
			}
			if ln, ok := locMap[t.LocationID]; ok {
				extra += fmt.Sprintf(" @%s", ln)
			}
			fmt.Printf("  %2d. %s%s%s\n", i+1, name, due, extra)
		}

		directDone, _ := cmd.Flags().GetInt("done")
		if directDone > 0 && directDone <= len(todos) {
			_, err := na.DoneTodo(ctx, todos[directDone-1].ID, "")
			if err != nil {
				return err
			}
			action.Printf("✅ 已完成: %s", todoNames[directDone-1])
			return nil
		}

		// Non-TTY mode (AI / piped): just list and exit, don't prompt
		if fi, _ := os.Stdin.Stat(); fi != nil && (fi.Mode()&os.ModeCharDevice) == 0 {
			fmt.Println("\n💡 非交互模式：使用 na daily --done N 完成待办")
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
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("读取输入失败: %w", err)
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
	action.Printf("🏠 已自动选择家庭: %s", families[0].Name)
	return nil
}

func init() {
	dailyCmd.Flags().Int("done", 0, "直接完成第 N 项待办")
}
