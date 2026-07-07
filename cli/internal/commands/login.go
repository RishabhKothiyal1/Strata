package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/client"
	"github.com/strata/strata/cli/internal/config"
	"github.com/strata/strata/cli/internal/output"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to a Strata project",
	Long: `Authenticate with the Strata API and store your access token.
You will be prompted for your email and password.`,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		if email == "" {
			fmt.Print("Email: ")
			fmt.Scanln(&email)
		}
		if password == "" {
			fmt.Print("Password: ")
			fmt.Scanln(&password)
		}

		cl := client.New()
		resp, err := cl.Login(email, password)
		if err != nil {
			output.Fatal(fmt.Errorf("login failed: %w", err))
		}

		cliCfg, err := config.LoadCLIConfig()
		if err != nil {
			output.Fatal(fmt.Errorf("failed to load config: %w", err))
		}

		projectCfg, _ := config.FindProjectConfig()
		projectName := "default"
		if projectCfg != nil {
			projectName = projectCfg.Name
		}

		cliCfg.CurrentProject = projectName
		if cliCfg.Projects == nil {
			cliCfg.Projects = make(map[string]*config.ProjectConfig)
		}
		cliCfg.Projects[projectName] = &config.ProjectConfig{
			Name:        projectName,
			AccessToken: resp.AccessToken,
			GatewayURL:  config.GetGatewayURL(),
		}

		if err := config.SaveCLIConfig(cliCfg); err != nil {
			output.Fatal(fmt.Errorf("failed to save config: %w", err))
		}

		output.Success("Logged in as %s", email)
		if user, ok := resp.User["role"]; ok {
			output.Info("Role: %s", user)
		}
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new account",
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		role, _ := cmd.Flags().GetString("role")
		orgID, _ := cmd.Flags().GetString("org")

		if email == "" {
			fmt.Print("Email: ")
			fmt.Scanln(&email)
		}
		if password == "" {
			fmt.Print("Password: ")
			fmt.Scanln(&password)
		}
		if role == "" {
			role = "user"
		}
		if orgID == "" {
			orgID = "default"
		}

		cl := client.New()
		if err := cl.Register(email, password, role, orgID); err != nil {
			output.Fatal(fmt.Errorf("registration failed: %w", err))
		}

		output.Success("Account created for %s", email)
		output.Info("Login with: strata login --email %s", email)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear stored credentials",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadCLIConfig()
		if err != nil {
			output.Fatal(fmt.Errorf("failed to load config: %w", err))
		}

		if cfg.CurrentProject != "" {
			delete(cfg.Projects, cfg.CurrentProject)
		}
		cfg.CurrentProject = ""
		config.SaveCLIConfig(cfg)

		output.Success("Logged out")
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current logged-in user",
	Run: func(cmd *cobra.Command, args []string) {
		token := config.GetAccessToken()
		if token == "" {
			output.Fatal(fmt.Errorf("not logged in - use 'strata login' first"))
		}

		cl := client.New()
		user, err := cl.GetMe()
		if err != nil {
			output.Fatal(fmt.Errorf("failed to get user info: %w", err))
		}

		output.Section("Current User")
		for k, v := range user {
			if k == "password_hash" {
				continue
			}
			fmt.Printf("  %-12s %v\n", k+":", v)
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("email", "", "Email address")
	loginCmd.Flags().String("password", "", "Password")

	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().String("email", "", "Email address")
	registerCmd.Flags().String("password", "", "Password")
	registerCmd.Flags().String("role", "user", "User role")
	registerCmd.Flags().String("org", "default", "Organization ID")

	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
}
