package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/client"
	"github.com/strata/strata/cli/internal/output"
)

type serviceCheck struct {
	Name string
	URL  string
}

var services = []serviceCheck{
	{"API Gateway", "http://localhost:8000"},
	{"Auth", "http://localhost:8081"},
	{"REST", "http://localhost:8082"},
	{"Realtime", "http://localhost:8083"},
	{"Storage", "http://localhost:8084"},
	{"Functions", "http://localhost:8085"},
	{"AI/Vector", "http://localhost:8086"},
	{"GraphQL", "http://localhost:8087"},
	{"PostgreSQL", "http://localhost:5432"},
	{"Redis", "http://localhost:6379"},
	{"MinIO", "http://localhost:9000"},
	{"NATS", "http://localhost:8222"},
	{"Prometheus", "http://localhost:9090"},
	{"Grafana", "http://localhost:3000"},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check health of all Strata services",
	Long: `Pings every Strata microservice and infrastructure component
to show which are running and responsive.`,
	Run: func(cmd *cobra.Command, args []string) {
		showAll, _ := cmd.Flags().GetBool("all")

		output.Info("Checking service health...")
		fmt.Println()

		cl := client.New()
		rows := [][]string{}
		allUp := true

		for _, svc := range services {
			isInfra := false
			for _, infra := range []string{"PostgreSQL", "Redis", "MinIO", "NATS", "Prometheus", "Grafana"} {
				if svc.Name == infra {
					isInfra = true
					break
				}
			}
			if isInfra && !showAll {
				continue
			}

			start := time.Now()
			_, err := cl.CheckServiceHealth(svc.URL)
			elapsed := time.Since(start)

			status := output.Green.Sprint("UP")
			latency := fmt.Sprintf("%dms", elapsed.Milliseconds())
			if err != nil {
				status = output.Red.Sprint("DOWN")
				latency = "-"
				allUp = false
			}
			rows = append(rows, []string{svc.Name, status, latency})
		}

		output.Table(
			[]string{"Service", "Status", "Latency"},
			rows,
		)

		if allUp {
			fmt.Println()
			output.Success("All services are running")
		} else {
			fmt.Println()
			output.Warn("Some services are unreachable - run 'strata start' to start them")
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolP("all", "a", false, "Include infrastructure services (PostgreSQL, Redis, etc.)")
}
