package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "strata",
	Short: "Strata - Open-source BaaS platform CLI",
	Long: `Strata is an enterprise-grade, open-source Backend-as-a-Service platform.

A central API Gateway manages authentication, rate limiting, and routing to
specialized services for REST APIs, GraphQL, realtime subscriptions, S3 storage,
serverless functions, and AI vector search.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Version: "1.0.0",
}

func Execute() {
	rootCmd.SetVersionTemplate("Strata CLI v{{.Version}}\n")

	cobra.AddTemplateFunc("green", func(s string) string {
		return color.GreenString(s)
	})
	rootCmd.SetHelpTemplate(getHelpTemplate())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is strata.json in current directory)")
	rootCmd.PersistentFlags().StringP("gateway", "g", "http://localhost:8000", "Strata API Gateway URL")
	rootCmd.PersistentFlags().StringP("token", "t", "", "Access token for authenticated requests")
}

func getHelpTemplate() string {
	return fmt.Sprintf(`%s

{{green "Usage:"}}
  {{.UseLine}}{{if .HasAvailableSubCommands}}

{{green "Available Commands:"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

{{if .HasAvailableLocalFlags}}{{green "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

{{if .HasAvailableInheritedFlags}}{{green "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

{{green "Learn more:"}}
  https://github.com/RishabhKothiyal1/novabase
`,
		color.CyanString("Strata CLI"),
	)
}
