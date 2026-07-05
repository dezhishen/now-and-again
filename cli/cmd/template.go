package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage task templates",
	Long:  `List templates and create tasks from them.`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		kind, _ := cmd.Flags().GetString("kind")
		templates, err := na.ListTemplates(context.Background(), kind)
		if err != nil {
			return err
		}
		if len(templates) == 0 {
			fmt.Println("No templates")
			return nil
		}
		for _, t := range templates {
			fmt.Printf("  %-30s  %-10s  %s  %s\n", t.TemplateCode, t.Kind, t.Icon, t.Name)
		}
		return nil
	},
}

var templateUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Create a task from a template",
	Long: `Renders a template with the given parameters and creates the task.

Example:
  na template use --code weekly_cleaning --params '{"area_name":"客厅"}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetString("code")
		paramsStr, _ := cmd.Flags().GetString("params")

		if code == "" {
			return fmt.Errorf("--code is required")
		}

		var params map[string]interface{}
		if paramsStr != "" {
			if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
				return fmt.Errorf("invalid --params JSON: %w", err)
			}
		}
		if params == nil {
			params = make(map[string]interface{})
		}

		t, err := na.CreateTaskFromTemplate(context.Background(), code, params)
		if err != nil {
			return err
		}
		fmt.Printf("Task created from template: %s (%s)\n", t.Name, t.ID[:8])
		return nil
	},
}

func init() {
	templateListCmd.Flags().String("kind", "", "Filter by kind (simple|inspection|chain)")
	templateUseCmd.Flags().String("code", "", "Template code (required)")
	templateUseCmd.Flags().String("params", "", "Template parameters as JSON")

	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateUseCmd)
}
