package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/output"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop local Strata services",
	Long: `Stops all Strata containers started by Docker Compose.
Use --volumes to also remove persistent data volumes.`,
	Run: func(cmd *cobra.Command, args []string) {
		volumes, _ := cmd.Flags().GetBool("volumes")

		if _, err := exec.LookPath("docker"); err != nil {
			output.Fatal(fmt.Errorf("docker is not installed or not in PATH"))
		}

		output.Info("Stopping Strata services...")

		argsList := []string{"compose", "down"}
		if volumes {
			argsList = append(argsList, "-v")
		}

		cmdProc := exec.Command("docker", argsList...)
		cmdProc.Stdout = os.Stdout
		cmdProc.Stderr = os.Stderr

		if err := cmdProc.Run(); err != nil {
			output.Fatal(fmt.Errorf("failed to stop services: %w", err))
		}

		output.Success("Services stopped")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolP("volumes", "v", false, "Remove persistent volumes (data loss)")
}
