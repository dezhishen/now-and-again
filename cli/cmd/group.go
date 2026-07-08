package cmd

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/dezhishen/now-and-again/cli/internal/resolver"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "小组管理",
	Long:  `创建、查看、加入和离开活跃家庭中的小组。`,
}

var groupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新小组",
	Example: `  na group create --name "大人" --desc "爸妈的任务"
  na group create --name "小孩"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name 不能为空")
		}
		desc, _ := cmd.Flags().GetString("desc")
		g, err := na.CreateGroup(context.Background(), name, desc)
		if err != nil {
			return err
		}
		action.Printf("✅ 小组已创建: %s (%s)", g.Name, g.ID[:8])
		return nil
	},
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出活跃家庭的所有小组",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		groups, err := na.ListGroups(context.Background(), "")
		if err != nil {
			return err
		}
		if len(groups) == 0 {
			action.Println("👥 暂无小组")
			return nil
		}
		fmt.Printf("👥 小组 (%d个):\n\n", len(groups))
		for _, g := range groups {
			fmt.Printf("  %-12s  %s\n", g.Name, g.ID[:8])
		}
		fmt.Println()
		return nil
	},
}

var groupJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "加入小组（支持名称或ID）",
	Example: `  na group join --name "大人"
  na group join --id abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetString("id")
		if name == "" && id == "" {
			return fmt.Errorf("请指定 --name 或 --id")
		}
		gid := id
		if gid == "" {
			cache := resolver.NewCache()
			var err error
			gid, err = cache.ResolveGroupID(context.Background(), na, name)
			if err != nil {
				return err
			}
		}
		if _, err := na.JoinGroup(context.Background(), gid); err != nil {
			return err
		}
		action.Println("✅ 已加入小组")
		return nil
	},
}

var groupLeaveCmd = &cobra.Command{
	Use:   "leave",
	Short: "离开小组（支持名称或ID）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetString("id")
		if name == "" && id == "" {
			return fmt.Errorf("请指定 --name 或 --id")
		}
		gid := id
		if gid == "" {
			cache := resolver.NewCache()
			var err error
			gid, err = cache.ResolveGroupID(context.Background(), na, name)
			if err != nil {
				return err
			}
		}
		if err := na.LeaveGroup(context.Background(), gid); err != nil {
			return err
		}
		action.Println("👋 已离开小组")
		return nil
	},
}

var groupMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "列出小组成员（支持名称或ID）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetString("id")
		if name == "" && id == "" {
			return fmt.Errorf("请指定 --name 或 --id")
		}
		gid := id
		if gid == "" {
			cache := resolver.NewCache()
			var err error
			gid, err = cache.ResolveGroupID(context.Background(), na, name)
			if err != nil {
				return err
			}
		}
		members, err := na.ListGroupMembers(context.Background(), gid)
		if err != nil {
			return err
		}
		if len(members) == 0 {
			action.Println("👥 暂无成员")
			return nil
		}
		fmt.Printf("👥 小组成员 (%d人):\n\n", len(members))
		for _, m := range members {
			fmt.Printf("  %s  %s\n", m.UserID[:8], m.Role)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	groupCreateCmd.Flags().String("name", "", "小组名称（必填）")
	groupCreateCmd.Flags().String("desc", "", "小组描述")

	groupJoinCmd.Flags().String("name", "", "小组名称")
	groupJoinCmd.Flags().String("id", "", "小组ID（完整UUID或≥3位前缀）")

	groupLeaveCmd.Flags().String("name", "", "小组名称")
	groupLeaveCmd.Flags().String("id", "", "小组ID（完整UUID或≥3位前缀）")

	groupMembersCmd.Flags().String("name", "", "小组名称")
	groupMembersCmd.Flags().String("id", "", "小组ID（完整UUID或≥3位前缀）")

	groupCmd.AddCommand(groupCreateCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupJoinCmd)
	groupCmd.AddCommand(groupLeaveCmd)
	groupCmd.AddCommand(groupMembersCmd)
}
