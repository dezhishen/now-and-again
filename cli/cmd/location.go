package cmd

import (
	"context"
	"fmt"

	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/dezhishen/now-and-again/cli/internal/resolver"
	"github.com/spf13/cobra"
)

var locationCmd = &cobra.Command{
	Use:   "location",
	Short: "地址管理",
	Long:  `创建、查看、更新和删除活跃家庭中的地址。`,
}

var locationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新地址",
	Example: `  na location create --name "厨房" --kind indoor
  na location create --name "阳台" --kind indoor --color "#4CAF50"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name 不能为空")
		}
		kind, _ := cmd.Flags().GetString("kind")
		if kind == "" {
			kind = "indoor"
		}
		color, _ := cmd.Flags().GetString("color")

		loc, err := na.CreateLocation(context.Background(), name, kind, color, "")
		if err != nil {
			return err
		}
		action.Printf("✅ 地址已创建: %s (%s)", loc.Name, loc.ID[:8])
		return nil
	},
}

var locationListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出活跃家庭的所有地址",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		locations, err := na.ListLocations(context.Background(), "")
		if err != nil {
			return err
		}
		if len(locations) == 0 {
			action.Println("📍 暂无地址")
			return nil
		}
		fmt.Printf("📍 地址 (%d个):\n\n", len(locations))
		for _, l := range locations {
			kind := l.Kind
			if kind == "" {
				kind = "indoor"
			}
			fmt.Printf("  %-12s  %-8s  %s\n", l.Name, kind, l.ID[:8])
		}
		fmt.Println()
		return nil
	},
}

var locationUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新地址属性",
	Example: `  na location update --name "厨房" --new-name "大厨房"
  na location update --id abc123 --color "#FF5722"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetString("id")
		if name == "" && id == "" {
			return fmt.Errorf("请指定 --name 或 --id")
		}

		locID := id
		if locID == "" {
			cache := resolver.NewCache()
			var err error
			locID, err = cache.ResolveLocationID(context.Background(), na, name)
			if err != nil {
				return err
			}
		}

		newName, _ := cmd.Flags().GetString("new-name")
		newKind, _ := cmd.Flags().GetString("new-kind")
		newColor, _ := cmd.Flags().GetString("new-color")

		loc, err := na.UpdateLocation(context.Background(), locID, newName, newKind, newColor)
		if err != nil {
			return err
		}
		action.Printf("✅ 已更新地址: %s", loc.Name)
		return nil
	},
}

var locationDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除地址",
	Example: `  na location delete --name "厨房"
  na location delete --id abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := autoEnsureFamily(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetString("id")
		if name == "" && id == "" {
			return fmt.Errorf("请指定 --name 或 --id")
		}

		locID := id
		if locID == "" {
			cache := resolver.NewCache()
			var err error
			locID, err = cache.ResolveLocationID(context.Background(), na, name)
			if err != nil {
				return err
			}
		}

		if err := na.DeleteLocation(context.Background(), locID); err != nil {
			return err
		}
		action.Printf("🗑️  已删除地址")
		return nil
	},
}

func init() {
	locationCreateCmd.Flags().String("name", "", "地址名称（必填）")
	locationCreateCmd.Flags().String("kind", "indoor", "地址类型（indoor/outdoor）")
	locationCreateCmd.Flags().String("color", "", "显示颜色（可选，如 #4CAF50）")

	locationUpdateCmd.Flags().String("name", "", "要更新的地址名称")
	locationUpdateCmd.Flags().String("id", "", "要更新的地址ID（完整UUID或≥3位前缀）")
	locationUpdateCmd.Flags().String("new-name", "", "新的地址名称")
	locationUpdateCmd.Flags().String("new-kind", "", "新的地址类型")
	locationUpdateCmd.Flags().String("new-color", "", "新的显示颜色")

	locationDeleteCmd.Flags().String("name", "", "地址名称")
	locationDeleteCmd.Flags().String("id", "", "地址ID（完整UUID或≥3位前缀）")

	locationCmd.AddCommand(locationCreateCmd)
	locationCmd.AddCommand(locationListCmd)
	locationCmd.AddCommand(locationUpdateCmd)
	locationCmd.AddCommand(locationDeleteCmd)
}
