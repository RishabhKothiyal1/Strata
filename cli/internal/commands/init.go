package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/config"
	"github.com/strata/strata/cli/internal/output"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Strata project",
	Long: `Creates a strata.json configuration file in the current directory.
This file tells the CLI how to connect to your Strata project.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "my-strata-project"
		if len(args) > 0 {
			name = args[0]
		}

		gatewayURL, _ := cmd.Flags().GetString("gateway")

		cfg := config.ProjectConfig{
			Name:       name,
			GatewayURL: gatewayURL,
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			output.Fatal(fmt.Errorf("failed to create config: %w", err))
		}

		path := filepath.Join(".", config.ConfigFileName)
		if _, err := os.Stat(path); err == nil {
			overwrite, _ := cmd.Flags().GetBool("force")
			if !overwrite {
				output.Fatal(fmt.Errorf("%s already exists (use --force to overwrite)", config.ConfigFileName))
			}
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			output.Fatal(fmt.Errorf("failed to write config: %w", err))
		}

		output.Success("Initialized project %s", name)
		output.Info("Created %s", config.ConfigFileName)
		fmt.Println()
		output.Dim.Println("  Gateway URL:", gatewayURL)
		fmt.Println()

		output.Section("Next steps")
		output.Info("Start local services:  strata start")
		output.Info("View service status:   strata status")
		output.Info("Login to your project: strata login")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing strata.json")
}
