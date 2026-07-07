package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/output"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage your project database",
	Long: `Manage database schemas, migrations, and seed data.
Ups streaming replication and schema introspection.`,
}

var dbPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local schema to the database",
	Long: `Push your schema.prisma or SQL migrations to the Strata database.
Creates tables, indexes, and applies any pending changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		schemaFile, _ := cmd.Flags().GetString("file")
		if schemaFile == "" {
			schemaFile = "schema.sql"
		}

		if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
			schemaDir := filepath.Join("docker", "postgres", "init")
			if entries, err := os.ReadDir(schemaDir); err == nil {
				for _, e := range entries {
					if filepath.Ext(e.Name()) == ".sql" {
						schemaFile = filepath.Join(schemaDir, e.Name())
						break
					}
				}
			}
			if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
				output.Fatal(fmt.Errorf("no schema file found (use --file to specify one)"))
			}
		}

		data, err := os.ReadFile(schemaFile)
		if err != nil {
			output.Fatal(fmt.Errorf("failed to read schema file: %w", err))
		}

		output.Info("Pushing schema from %s...", schemaFile)

		psqlCmd := exec.Command("psql",
			"-h", "localhost",
			"-U", "strata_admin",
			"-d", "strata",
			"-f", schemaFile,
		)
		psqlCmd.Env = append(os.Environ(), "PGPASSWORD=strata_secure_pass_123")
		psqlCmd.Stdout = os.Stdout
		psqlCmd.Stderr = os.Stderr

		if err := psqlCmd.Run(); err != nil {
			output.Fatal(fmt.Errorf("failed to push schema: %w", err))
		}

		output.Success("Schema applied successfully")
		output.Info("Tables are now available via the REST API at /v1/rest/")
	},
}

var dbPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull remote schema to a local file",
	Long: `Introspects the Strata PostgreSQL database and writes
the current schema to a local SQL file.`,
	Run: func(cmd *cobra.Command, args []string) {
		outFile, _ := cmd.Flags().GetString("out")
		if outFile == "" {
			outFile = "schema.sql"
		}

		output.Info("Introspecting database schema...")

		pgDump := exec.Command("pg_dump",
			"-h", "localhost",
			"-U", "strata_admin",
			"-d", "strata",
			"--schema-only",
			"-f", outFile,
		)
		pgDump.Env = append(os.Environ(), "PGPASSWORD=strata_secure_pass_123")
		pgDump.Stderr = os.Stderr

		if err := pgDump.Run(); err != nil {
			output.Fatal(fmt.Errorf("failed to pull schema: %w", err))
		}

		output.Success("Schema written to %s", outFile)
	},
}

var dbSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with sample data",
	Run: func(cmd *cobra.Command, args []string) {
		seedFile, _ := cmd.Flags().GetString("file")
		if seedFile == "" {
			seedFile = "seed.sql"
		}

		if _, err := os.Stat(seedFile); os.IsNotExist(err) {
			output.Warn("No seed file found at %s", seedFile)
			return
		}

		output.Info("Seeding database...")

		psqlCmd := exec.Command("psql",
			"-h", "localhost",
			"-U", "strata_admin",
			"-d", "strata",
			"-f", seedFile,
		)
		psqlCmd.Env = append(os.Environ(), "PGPASSWORD=strata_secure_pass_123")
		psqlCmd.Stdout = os.Stdout
		psqlCmd.Stderr = os.Stderr

		if err := psqlCmd.Run(); err != nil {
			output.Fatal(fmt.Errorf("failed to seed database: %w", err))
		}

		output.Success("Database seeded successfully")
	},
}

func init() {
	dbCmd.AddCommand(dbPushCmd)
	dbCmd.AddCommand(dbPullCmd)
	dbCmd.AddCommand(dbSeedCmd)

	dbPushCmd.Flags().StringP("file", "f", "", "Path to SQL schema file")
	dbPullCmd.Flags().StringP("out", "o", "schema.sql", "Output file path")
	dbSeedCmd.Flags().StringP("file", "f", "seed.sql", "Path to seed SQL file")

	rootCmd.AddCommand(dbCmd)
}
