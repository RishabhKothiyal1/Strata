package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/client"
	"github.com/strata/strata/cli/internal/output"
)

var functionsCmd = &cobra.Command{
	Use:     "functions",
	Aliases: []string{"fn", "func"},
	Short:   "Manage serverless functions",
	Long:    `Deploy, list, invoke, and manage serverless JavaScript functions.`,
}

var functionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployed functions",
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		fns, err := cl.ListFunctions()
		if err != nil {
			output.Fatal(fmt.Errorf("failed to list functions: %w", err))
		}
		if len(fns) == 0 {
			output.Info("No functions deployed")
			output.Info("Deploy one with: strata functions deploy <name> <file>")
			return
		}
		rows := [][]string{}
		for _, fn := range fns {
			rows = append(rows, []string{fn.Name, fn.Description, fn.CreatedAt})
		}
		output.Table([]string{"Name", "Description", "Created"}, rows)
	},
}

var functionsDeployCmd = &cobra.Command{
	Use:   "deploy <name> <file>",
	Short: "Deploy a new function",
	Long: `Deploy a JavaScript function to Strata.
The file should export a handler function:

  function handler(request) {
    return { statusCode: 200, body: { message: "Hello!" } };
  }`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		filePath := args[1]

		data, err := os.ReadFile(filePath)
		if err != nil {
			output.Fatal(fmt.Errorf("failed to read file: %w", err))
		}

		description, _ := cmd.Flags().GetString("description")

		cl := client.New()
		if err := cl.DeployFunction(name, description, string(data)); err != nil {
			output.Fatal(fmt.Errorf("deploy failed: %w", err))
		}

		output.Success("Function '%s' deployed", name)
		output.Info("Invoke with: strata functions invoke %s", name)
	},
}

var functionsInvokeCmd = &cobra.Command{
	Use:   "invoke <name> [payload]",
	Short: "Invoke a deployed function",
	Long: `Invoke a function with an optional JSON payload.
If no payload is given, an empty body is sent.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		var payload interface{}
		payloadStr, _ := cmd.Flags().GetString("data")
		if payloadStr != "" {
			if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
				payload = map[string]interface{}{"body": payloadStr}
			}
		} else {
			payload = map[string]interface{}{}
		}

		cl := client.New()
		result, err := cl.InvokeFunction(name, payload)
		if err != nil {
			output.Fatal(fmt.Errorf("invocation failed: %w", err))
		}

		output.Success("Function '%s' returned:", name)
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	},
}

var functionsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a deployed function",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cl := client.New()
		if err := cl.DeleteFunction(name); err != nil {
			output.Fatal(fmt.Errorf("delete failed: %w", err))
		}
		output.Success("Function '%s' deleted", name)
	},
}

func init() {
	functionsCmd.AddCommand(functionsListCmd)
	functionsCmd.AddCommand(functionsDeployCmd)
	functionsCmd.AddCommand(functionsInvokeCmd)
	functionsCmd.AddCommand(functionsDeleteCmd)

	functionsDeployCmd.Flags().StringP("description", "d", "", "Function description")
	functionsInvokeCmd.Flags().StringP("data", "d", "", "JSON payload to send")

	rootCmd.AddCommand(functionsCmd)
}
