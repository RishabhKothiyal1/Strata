package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/output"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start local Strata services with Docker Compose",
	Long: `Starts all Strata microservices and infrastructure using Docker Compose.
By default it runs in detached mode.`,
	Run: func(cmd *cobra.Command, args []string) {
		detach, _ := cmd.Flags().GetBool("detach")
		build, _ := cmd.Flags().GetBool("build")

		if _, err := exec.LookPath("docker"); err != nil {
			output.Fatal(fmt.Errorf("docker is not installed or not in PATH"))
		}
		if _, err := exec.LookPath("docker-compose"); err != nil {
			if _, err := exec.LookPath("docker"); err != nil {
				output.Fatal(fmt.Errorf("docker compose plugin not found"))
			}
		}

		output.Info("Starting Strata services...")

		argsList := []string{"compose", "up"}
		if detach {
			argsList = append(argsList, "-d")
		}
		if build {
			argsList = append(argsList, "--build")
		}

		cmdProc := exec.Command("docker", argsList...)
		cmdProc.Stdout = os.Stdout
		cmdProc.Stderr = os.Stderr

		if err := cmdProc.Run(); err != nil {
			output.Fatal(fmt.Errorf("failed to start services: %w", err))
		}

		if detach {
			output.Success("Services started in detached mode")
			fmt.Println()
			output.Section("Running services")
			statusCmd.Run(cmd, args)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("detach", "d", true, "Run containers in detached mode")
	startCmd.Flags().BoolP("build", "b", false, "Rebuild images before starting")
}
